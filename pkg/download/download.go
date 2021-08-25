package download

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/database"
	"github.com/jmoiron/sqlx"
)

type Downloader struct {
	RateLimitDuration time.Duration
	Config            config.Config
}

type Download struct {
	URL  string
	Etag string
	Size int
}

func NewDownloader(cfg config.Config) *Downloader {
	return &Downloader{
		Config: cfg,
	}
}

func (d Downloader) FileInCache(fullurl string) (bool, error) {
	parsed_url, err := url.Parse(fullurl)
	if err != nil {
		return false, nil
	}

	filePath := filepath.Join(d.Config.Main.CacheDir, parsed_url.Path)

	filestat, err := os.Stat(filePath)
	if err != nil {
		return false, nil
	}

	db, err := database.ConnectDB(d.Config)
	if err != nil {
		return false, err
	}

	is_consistent, err := d.FileConsistent(db, filestat, fullurl)
	if err != nil {
		return false, err
	}

	if filestat != nil && !is_consistent {
		return false, err
	}

	return true, nil
}

func (d Downloader) FileConsistent(db *sqlx.DB, file fs.FileInfo, fullurl string) (bool, error) {
	var downloads []Download

	err := db.Select(&downloads, "SELECT url, etag, size FROM sec.downloads WHERE url = $1", fullurl)
	if err != nil {
		return false, err
	}

	if len(downloads) == 0 {
		return false, nil
	}

	req, err := http.NewRequest("HEAD", fullurl, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)")

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return false, err
	}

	etag := resp.Header.Get("eTag")
	if err != nil {
		return false, err
	}

	var download Download
	if len(downloads) > 0 {
		download = downloads[0]
	}

	if download.Etag != etag {
		return false, nil
	}

	return true, nil
}

func (d Downloader) DownloadFile(db *sqlx.DB, fullurl string) error {
	fileUrl, err := url.Parse(fullurl)
	if err != nil {
		return err
	}
	cachePath := filepath.Join(d.Config.Main.CacheDir, fileUrl.Path)

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	etag := resp.Header.Get("eTag")

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	size, err := io.Copy(ioutil.Discard, bytes.NewReader(responseBody))
	if err != nil {
		return err
	}

	foldersPath := strings.ReplaceAll(cachePath, filepath.Base(cachePath), "")
	if _, err = os.Stat(foldersPath); err != nil {
		err = os.MkdirAll(foldersPath, 0755)
		if err != nil {
			return err
		}
	}

	out, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(responseBody))
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO sec.downloads (url, etag, size, created_at, updated_at) 
	VALUES ($1, $2, $3, NOW(), NOW()) 
	ON CONFLICT (url) 
	DO UPDATE SET url=EXCLUDED.url, etag=EXCLUDED.etag, size=EXCLUDED.size, updated_at=NOW() 
	WHERE downloads.url=EXCLUDED.url;`, fullurl, etag, size)
	if err != nil {
		return err
	}

	return nil
}
