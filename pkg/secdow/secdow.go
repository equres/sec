package secdow

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

func DownloadTickerFile(db *sqlx.DB, s *sec.SEC, path string) error {
	downloader := download.NewDownloader(s.Config)
	downloader.IsEtag = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = 1

	baseURL, err := url.Parse(s.BaseURL)
	if err != nil {
		return err
	}

	pathURL, err := url.Parse(path)
	if err != nil {
		return err
	}

	fullURL := baseURL.ResolveReference(pathURL).String()

	s.Log(fmt.Sprintf("Checking for file %v: ", filepath.Base(pathURL.Path)))

	etag, err := downloader.GetFileETag(fullURL)
	if err != nil {
		return err
	}

	isFileCorrect, err := downloader.FileCorrect(db, fullURL, 0, etag)
	if err != nil {
		return err
	}

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	if isFileCorrect {
		s.Log("\u2713")
	}
	if !isFileCorrect {
		s.Log("Downloading file...: ")
		fileDowStartTime := time.Now()
		err = downloader.DownloadFile(db, fullURL)
		if err != nil {
			return err
		}
		timeTaken := time.Since(fileDowStartTime)
		timeToWait := rateLimit - timeTaken

		if timeToWait.Milliseconds() > 0 {
			time.Sleep(time.Duration(timeToWait.Milliseconds()))
		}

		s.Log(time.Now().Format("2006-01-02 03:04:05"))
	}
	return nil
}

func DownloadIndex(db *sqlx.DB, s *sec.SEC) error {
	worklist, err := secworklist.WillDownloadGet(db, true)
	if err != nil {
		return err
	}

	downloader := download.NewDownloader(s.Config)
	downloader.IsEtag = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = len(worklist)

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	for _, v := range worklist {
		fileURL, err := secutil.FormatFilePathDate(s.BaseURL, v.Year, v.Month)
		if err != nil {
			return err
		}

		s.Log(fmt.Sprintf("Checking file '%v' in disk: ", filepath.Base(fileURL)))

		etag, err := downloader.GetFileETag(fileURL)
		if err != nil {
			return err
		}

		isFileCorrect, err := downloader.FileCorrect(db, fileURL, 0, etag)
		if err != nil {
			return err
		}
		if isFileCorrect {
			s.Log("\u2713")
		}

		if !isFileCorrect {
			s.Log("Downloading file...: ")

			fileDowStartTime := time.Now()
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			timeTaken := time.Since(fileDowStartTime)
			timeToWait := rateLimit - timeTaken

			if timeToWait.Milliseconds() > 0 {
				time.Sleep(time.Duration(timeToWait.Milliseconds()))
			}

			s.Log(time.Now().Format("2006-01-02 03:04:05"))
		}

		downloader.CurrentDownloadCount += 1
	}
	return nil
}

func DownloadXbrlFileContent(db *sqlx.DB, s *sec.SEC, files []sec.XbrlFile, config config.Config, currentCount *int, totalCount int) error {
	downloader := download.NewDownloader(s.Config)
	downloader.IsEtag = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.CurrentDownloadCount = *currentCount
	downloader.TotalDownloadsCount = totalCount

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	for _, v := range files {
		size, err := strconv.Atoi(v.Size)
		if err != nil {
			return err
		}
		isFileCorrect, err := downloader.FileCorrect(db, v.URL, size, "")
		if err != nil {
			return err
		}

		if !isFileCorrect {
			fileDowStartTime := time.Now()
			err = downloader.DownloadFile(db, v.URL)
			if err != nil {
				return err
			}
			timeTaken := time.Since(fileDowStartTime)
			timeToWait := rateLimit - timeTaken

			if timeToWait.Milliseconds() > 0 {
				time.Sleep(time.Duration(timeToWait.Milliseconds()))
			}
		}

		*currentCount++
		downloader.CurrentDownloadCount = *currentCount
		if !s.Verbose {
			log.Info(fmt.Sprintf("\r[%d/%d/%f%% files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", *currentCount, totalCount, downloader.GetDownloadPercentage(), (totalCount - *currentCount)))
		}

		s.Log(fmt.Sprintf("[%d/%d/%f%%] %s downloaded...\n", *currentCount, totalCount, downloader.GetDownloadPercentage(), time.Now().Format("2006-01-02 03:04:05")))
		time.Sleep(rateLimit)
	}

	return nil
}

func DownloadAllItemFiles(db *sqlx.DB, s *sec.SEC, rssFile sec.RSSFile, worklist []secworklist.Worklist) error {
	s.Log("Calculating number of XBRL Files in the index files: ")

	totalCount, err := secutil.TotalXbrlFileCountGet(worklist, s, s.Config.Main.CacheDir)
	if err != nil {
		return err
	}

	s.Log(fmt.Sprint(totalCount))

	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		err := DownloadXbrlFileContent(db, s, v1.XbrlFiling.XbrlFiles.XbrlFile, s.Config, &currentCount, totalCount)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadZIPFiles(db *sqlx.DB, s *sec.SEC) error {
	downloader := download.NewDownloader(s.Config)
	downloader.IsContentLength = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	worklist, err := secworklist.WillDownloadGet(db, true)
	if err != nil {
		return err
	}

	totalCount, err := secutil.GetTotalZIPFilesToBeDownloaded(db, s, worklist)
	if err != nil {
		return err
	}
	currentCount := 0

	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = totalCount
	for _, v := range worklist {
		fileURL, err := secutil.FormatFilePathDate(s.Config.Main.CacheDir, v.Year, v.Month)
		if err != nil {
			return err
		}

		_, err = os.Stat(fileURL)
		if err != nil {
			return fmt.Errorf("please run sec dow index to download all index files first")
		}

		rssFile, err := secutil.ParseRSSGoXML(fileURL)
		if err != nil {
			return err
		}

		for _, v1 := range rssFile.Channel.Item {
			if v1.Enclosure.URL != "" {
				size, err := strconv.Atoi(v1.Enclosure.Length)
				if err != nil {
					return err
				}

				isFileCorrect, err := downloader.FileCorrect(db, v1.Enclosure.URL, size, "")
				if err != nil {
					return err
				}

				if !isFileCorrect {
					fileDowStartTime := time.Now()
					err = downloader.DownloadFile(db, v1.Enclosure.URL)
					if err != nil {
						return err
					}
					timeTaken := time.Since(fileDowStartTime)
					timeToWait := rateLimit - timeTaken

					if timeToWait.Milliseconds() > 0 {
						time.Sleep(time.Duration(timeToWait.Milliseconds()))
					}
				}
			}
			currentCount++
			downloader.CurrentDownloadCount = currentCount
			if !s.Verbose {
				log.Info(fmt.Sprintf("\r[%d/%d/%f%% files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", currentCount, totalCount, downloader.GetDownloadPercentage(), (totalCount - currentCount)))
			}

			s.Log(fmt.Sprintf("[%d/%d/%f%%] %s downloaded...\n", currentCount, totalCount, downloader.GetDownloadPercentage(), time.Now().Format("2006-01-02 03:04:05")))
		}

	}
	return nil
}

func DownloadRawFiles(s *sec.SEC, db *sqlx.DB, filesToDownload []string, totalDownloadsCount int, currentDownloadCount int) error {
	s.Log(fmt.Sprint("Number of files to be downloaded: ", totalDownloadsCount))

	downloader := download.NewDownloader(s.Config)
	downloader.IsEtag = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.TotalDownloadsCount = totalDownloadsCount
	downloader.CurrentDownloadCount = currentDownloadCount

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	startTime := time.Now()
	averageDownloadTime := time.Duration(0)
	for k, v := range filesToDownload {
		s.Log(fmt.Sprintf("Download progress [%d/%d/%f%%] Time To Complete All Downloads: %v", downloader.CurrentDownloadCount, downloader.TotalDownloadsCount, downloader.GetDownloadPercentage(), averageDownloadTime))

		fileDowStartTime := time.Now()
		err = downloader.DownloadFile(db, v)
		if err != nil {
			return err
		}
		timeTaken := time.Since(fileDowStartTime)
		timeToWait := rateLimit - timeTaken

		if timeToWait.Milliseconds() > 0 {
			time.Sleep(time.Duration(timeToWait.Milliseconds()))
		}
		if k != 0 && k%1000 == 0 {
			s.Log(fmt.Sprint("The past 1000 files took around ", time.Since(startTime), " to download, this means each file took ", time.Since(startTime)/1000, " to download"))
			filesRemaining := (downloader.TotalDownloadsCount - downloader.CurrentDownloadCount)
			averageDownloadTime = time.Duration((time.Since(startTime) / 1000).Nanoseconds() * int64(filesRemaining))
			startTime = time.Now()
		}
		downloader.CurrentDownloadCount += 1
	}

	return nil
}
