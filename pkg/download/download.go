package download

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secreq"
	"github.com/jmoiron/sqlx"
)

type Downloader struct {
	RateLimitDuration    time.Duration
	Config               config.Config
	Verbose              bool
	Debug                bool
	IsEtag               bool
	IsContentLength      bool
	CurrentDownloadCount int
	TotalDownloadsCount  int
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

func (d Downloader) FileCorrect(db *sqlx.DB, fullurl string, size int, etag string) (bool, error) {
	parsedURL, err := url.Parse(fullurl)
	if err != nil {
		return false, err
	}

	isFileInCache, err := d.FileInCache(filepath.Join(d.Config.Main.CacheDir, parsedURL.Path))
	if err != nil {
		if d.Verbose {
			log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] not_in_cache", parsedURL, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))
		}
		return false, nil
	}

	if isFileInCache == nil {
		if d.Verbose {
			log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] not_in_cache", parsedURL, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))
		}
		return false, nil
	}

	isConsistent, err := d.FileConsistent(db, isFileInCache, fullurl, size, etag)
	if err != nil {
		return false, err
	}

	if isFileInCache != nil && !isConsistent {
		if d.Verbose {
			log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] in_cache_not_consistent", parsedURL, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))
		}
		return false, nil
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

func (d Downloader) GetFileETag(fullURL string) (string, error) {
	req := secreq.NewSECReqHEAD(d.Config)
	req.IsEtag = true
	req.IsContentLength = false

	retryLimit, err := strconv.Atoi(d.Config.Main.RetryLimit)
	if err != nil {
		return "", err
	}

	rateLimit, err := strconv.Atoi(d.Config.Main.RateLimitMs)
	if err != nil {
		return "", err
	}

	resp, err := req.SendRequest(retryLimit, time.Duration(rateLimit), fullURL)
	if err != nil {
		return "", err
	}

	etag := resp.Header.Get("eTag")

	return etag, nil
}

func (d Downloader) FileConsistent(db *sqlx.DB, file fs.FileInfo, fullurl string, size int, etag string) (bool, error) {
	var downloads []Download
	var err error

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
		if d.Verbose {
			log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] no_download_in_the_database", fullurl, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))
		}
		return false, nil
	}

	download := downloads[0]

	if d.IsEtag && (etag != "" && download.Etag == etag) {
		return true, nil
	}

	if download.Size != size {
		log.Info("Expected Size:", size, "vs Actual File Size:", download.Size)
		return false, nil
	}

	return true, nil
}

func (d Downloader) DownloadFile(db *sqlx.DB, fullurl string) error {
	fileDowStartTime := time.Now()

	isSkippedFile, err := database.IsSkippedFile(db, fullurl)
	if err != nil {
		return err
	}

	if isSkippedFile {
		log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] skipped_downloading", fullurl, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))
		return nil
	}

	log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] currently_downloading", fullurl, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage()))

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

	req := secreq.NewSECReqGET(d.Config)
	req.IsEtag = d.IsEtag
	req.IsContentLength = d.IsContentLength

	resp, err := req.SendRequest(retryLimit, rateLimit, fullurl)
	if err != nil {
		secevent.CreateDownloadEvent(db, cachePath, fullurl, "failed", err.Error())

		if err.Error() == errors.New("404").Error() {
			insertErr := database.SkipFileInsert(db, fullurl)
			if insertErr != nil {
				return insertErr
			}
			return nil
		}
	}

	if d.Debug {
		log.Info()
		headers, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}
		log.Info(string(headers))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] status_code_%d", fullurl, d.CurrentDownloadCount, d.TotalDownloadsCount, d.GetDownloadPercentage(), resp.StatusCode))
	if IsErrorPage(string(responseBody)) {
		secevent.CreateDownloadEvent(db, cachePath, fullurl, "failed", fmt.Sprintf("returned error page - Status Code: %v", resp.StatusCode))
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

	secevent.CreateDownloadEvent(db, cachePath, fullurl, "success", "")

	if d.IsEtag {
		etag := resp.Header.Get("eTag")
		err = IndexEtag(*db, fullurl, etag, size)
		if err != nil {
			return err
		}
	} else {
		err = IndexContentLength(*db, fullurl, size)
		if err != nil {
			return err
		}
	}

	timeTaken := time.Since(fileDowStartTime)
	timeToWait := rateLimit - timeTaken

	if timeToWait.Milliseconds() > 0 {
		time.Sleep(time.Duration(timeToWait.Milliseconds()))
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

func (d Downloader) GetDownloadPercentage() float64 {
	currentCountFloat := float64(d.CurrentDownloadCount)
	totalCountFloat := float64(d.TotalDownloadsCount)

	return (currentCountFloat / totalCountFloat) * 100
}
