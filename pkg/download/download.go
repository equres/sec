package download

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/secreq"
	"github.com/jmoiron/sqlx"
)

type Downloader struct {
	RateLimitDuration time.Duration
	Config            config.Config
	Verbose           bool
	Debug             bool
	IsEtag            bool
	IsContentLength   bool
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

func (d Downloader) FileCorrect(db *sqlx.DB, fullurl string) (bool, error) {
	parsedURL, err := url.Parse(fullurl)
	if err != nil {
		return false, err
	}

	isFileInCache, err := d.FileInCache(filepath.Join(d.Config.Main.CacheDir, parsedURL.Path))
	if err != nil {
		if d.Verbose {
			log.Print("File is not in cache: ")
		}
		return false, nil
	}

	isConsistent, err := d.FileConsistent(db, isFileInCache, fullurl)
	if err != nil {
		return false, err
	}

	if isFileInCache != nil && !isConsistent {
		if d.Verbose {
			log.Print("File in cache not consistent: ")
		}
		return false, err
	}

	return true, nil
}

func (d Downloader) FileInCache(path string) (fs.FileInfo, error) {
	filestat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return filestat, nil
}

func (d Downloader) FileConsistent(db *sqlx.DB, file fs.FileInfo, fullurl string) (bool, error) {
	var downloads []Download
	retryLimit, err := strconv.Atoi(d.Config.Main.RetryLimit)
	if err != nil {
		return false, err
	}

	rateLimit, err := time.ParseDuration(d.Config.Main.RateLimitMs + "ms")
	if err != nil {
		return false, err
	}

	if d.IsEtag {
		err = db.Select(&downloads, "SELECT url, etag, size FROM sec.downloads WHERE url = $1", fullurl)
		if err != nil {
			return false, err
		}
	}
	if d.IsContentLength {
		err = db.Select(&downloads, "SELECT url, size FROM sec.downloads WHERE url = $1", fullurl)
		if err != nil {
			return false, err
		}
	}

	if len(downloads) == 0 {
		return false, nil
	}

	var download Download
	if len(downloads) > 0 {
		download = downloads[0]
	}

	req := secreq.NewSECReqHEAD()
	req.IsEtag = d.IsEtag
	req.IsContentLength = d.IsContentLength

	resp, err := req.SendRequest(retryLimit, rateLimit, fullurl)
	if err != nil {
		return false, err
	}

	if d.Debug {
		log.Println()
		headers, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return false, err
		}
		log.Println(string(headers))
	}

	if d.IsEtag {
		etag := resp.Header.Get("eTag")
		if download.Etag != etag {
			return false, nil
		}
	}

	if d.IsContentLength {
		contentLengthHeader := resp.Header.Get("Content-Length")
		contentLength, err := strconv.Atoi(contentLengthHeader)
		if err != nil {
			return false, err
		}
		if download.Size != contentLength {
			return false, nil
		}
	}

	return true, nil
}

func (d Downloader) DownloadFile(db *sqlx.DB, fullurl string) error {
	retryLimit, err := strconv.Atoi(d.Config.Main.RetryLimit)
	if err != nil {
		return err
	}

	fileUrl, err := url.Parse(fullurl)
	if err != nil {
		return err
	}
	cachePath := filepath.Join(d.Config.Main.CacheDir, fileUrl.Path)

	rateLimit, err := time.ParseDuration(d.Config.Main.RateLimitMs + "ms")
	if err != nil {
		return err
	}

	req := secreq.NewSECReqGET()
	req.IsEtag = d.IsEtag
	req.IsContentLength = d.IsContentLength

	resp, err := req.SendRequest(retryLimit, rateLimit, fullurl)
	if err != nil {
		return err
	}

	if d.Debug {
		log.Println()
		headers, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}
		log.Println(string(headers))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println("Status Code:", resp.StatusCode)
	if IsErrorPage(string(responseBody)) {
		return fmt.Errorf("requested file but received an error instead")
	}

	size, err := io.Copy(ioutil.Discard, bytes.NewReader(responseBody))
	if err != nil {
		return err
	}

	err = SaveFile(cachePath, responseBody)
	if err != nil {
		return err
	}

	if d.IsEtag {
		etag := resp.Header.Get("eTag")
		err = IndexEtag(*db, fullurl, etag, size)
		if err != nil {
			return err
		}
	}

	if d.IsContentLength {
		err = IndexContentLength(*db, fullurl, size)
		if err != nil {
			return err
		}
	}

	return nil
}

func IsErrorPage(data string) bool {
	return strings.Contains(data, "This page is temporarily unavailable.")
}

func SaveFile(cachePath string, responseBody []byte) error {
	foldersPath := strings.ReplaceAll(cachePath, filepath.Base(cachePath), "")
	if _, err := os.Stat(foldersPath); err != nil {
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

	return nil
}

func IndexEtag(db sqlx.DB, fullurl string, etag string, size int64) error {
	_, err := db.Exec(`
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

func IndexContentLength(db sqlx.DB, fullurl string, size int64) error {
	_, err := db.Exec(`
		INSERT INTO sec.downloads (url, size, created_at, updated_at) 
		VALUES ($1, $2, NOW(), NOW()) 
		ON CONFLICT (url) 
		DO UPDATE SET url=EXCLUDED.url, size=EXCLUDED.size, updated_at=NOW() 
		WHERE downloads.url=EXCLUDED.url;`, fullurl, size)
	if err != nil {
		return err
	}

	return nil
}
