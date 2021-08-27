package download

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
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

func (d Downloader) FileInCache(db *sqlx.DB, fullurl string) (bool, error) {
	parsed_url, err := url.Parse(fullurl)
	if err != nil {
		return false, nil
	}

	filePath := filepath.Join(d.Config.Main.CacheDir, parsed_url.Path)

	filestat, err := os.Stat(filePath)
	if err != nil {
		if d.Verbose {
			fmt.Print("File is not in cache: ")
		}
		return false, nil
	}

	is_consistent, err := d.FileConsistent(db, filestat, fullurl)
	if err != nil {
		return false, err
	}

	if filestat != nil && !is_consistent {
		if d.Verbose {
			fmt.Print("File in cache not consistent: ")
		}
		return false, err
	}

	return true, nil
}

func (d Downloader) FileConsistent(db *sqlx.DB, file fs.FileInfo, fullurl string) (bool, error) {
	var downloads []Download
	retryCountStr := d.Config.Main.RetryLimit
	retryCount, err := strconv.Atoi(retryCountStr)
	if err != nil {
		return false, err
	}
	currentRetryCount := retryCount

	rateLimit, err := time.ParseDuration(d.Config.Main.RateLimitMs + "ms")
	if err != nil {
		return false, err
	}

	err = db.Select(&downloads, "SELECT url, etag, size FROM sec.downloads WHERE url = $1", fullurl)
	if err != nil {
		return false, err
	}

	if len(downloads) == 0 {
		return false, nil
	}

	var req *http.Request
	var resp *http.Response
	var etag string

	for currentRetryCount > 0 {
		currentRetryCount--
		req, err = secreq.NewSECReqHEAD(fullurl)
		if err != nil {
			return false, err
		}

		resp, err = new(http.Client).Do(req)
		if err != nil {
			return false, err
		}
		etag = resp.Header.Get("eTag")

		if d.Debug {
			fmt.Println()
			headers, err := httputil.DumpResponse(resp, false)
			if err != nil {
				return false, err
			}
			fmt.Print(string(headers))
		}

		if etag != "" {
			break
		}
		if d.Debug && currentRetryCount == retryCount-1 {
			fmt.Print("HEAD Request failed, retrying...: ")
		}
		time.Sleep(rateLimit)
	}

	if currentRetryCount == 0 && etag == "" {
		return false, fmt.Errorf("retried to retrieve headers and failed %v times", d.Config.Main.RetryLimit)
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
	retryCountStr := d.Config.Main.RetryLimit
	retryCount, err := strconv.Atoi(retryCountStr)
	if err != nil {
		return err
	}
	currentRetryCount := retryCount

	fileUrl, err := url.Parse(fullurl)
	if err != nil {
		return err
	}
	cachePath := filepath.Join(d.Config.Main.CacheDir, fileUrl.Path)

	rateLimit, err := time.ParseDuration(d.Config.Main.RateLimitMs + "ms")
	if err != nil {
		return err
	}

	client := &http.Client{}

	var req *http.Request
	var resp *http.Response
	var etag string

	for currentRetryCount > 0 {
		currentRetryCount--
		req, err = secreq.NewSECReqGET(fullurl)
		if err != nil {
			return err
		}

		resp, err = client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		etag = resp.Header.Get("eTag")

		if d.Debug {
			fmt.Println()
			headers, err := httputil.DumpResponse(resp, false)
			if err != nil {
				return err
			}
			fmt.Println(string(headers))
		}

		if etag != "" {
			break
		}

		time.Sleep(rateLimit)
	}

	if currentRetryCount == 0 && etag == "" {
		return fmt.Errorf("retried to download file and fail %v times", d.Config.Main.RetryLimit)
	}

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
