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

	if s.Verbose {
		log.Info(fmt.Sprintf("Checking for file %v: ", filepath.Base(pathURL.Path)))
	}

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

	if s.Verbose && isFileCorrect {
		log.Info("\u2713")
	}
	if !isFileCorrect {
		if s.Verbose {
			log.Info("Downloading file...: ")
		}
		err = downloader.DownloadFile(db, fullURL)
		if err != nil {
			return err
		}

		if s.Verbose {
			log.Info(time.Now().Format("2006-01-02 03:04:05"))
		}
		time.Sleep(rateLimit)
	}
	return nil
}

func DownloadIndex(db *sqlx.DB, s *sec.SEC) error {
	worklist, err := secworklist.WillDownloadGet(db)
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

		if s.Verbose {
			log.Info(fmt.Sprintf("Checking file '%v' in disk: ", filepath.Base(fileURL)))
		}

		etag, err := downloader.GetFileETag(fileURL)
		if err != nil {
			return err
		}

		isFileCorrect, err := downloader.FileCorrect(db, fileURL, 0, etag)
		if err != nil {
			return err
		}
		if s.Verbose && isFileCorrect {
			log.Info("\u2713")
		}

		if !isFileCorrect {
			if s.Verbose {
				log.Info("Downloading file...: ")
			}

			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			if s.Verbose {
				log.Info(time.Now().Format("2006-01-02 03:04:05"))
			}
			time.Sleep(rateLimit)
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
			err = downloader.DownloadFile(db, v.URL)
			if err != nil {
				return err
			}
			time.Sleep(rateLimit)
		}

		*currentCount++
		downloader.CurrentDownloadCount = *currentCount
		if !s.Verbose {
			log.Info(fmt.Sprintf("\r[%d/%d/%f%% files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", *currentCount, totalCount, downloader.GetDownloadPercentage(), (totalCount - *currentCount)))
		}

		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d/%f%%] %s downloaded...\n", *currentCount, totalCount, downloader.GetDownloadPercentage(), time.Now().Format("2006-01-02 03:04:05")))
		}
		time.Sleep(rateLimit)
	}

	return nil
}

func DownloadAllItemFiles(db *sqlx.DB, s *sec.SEC, rssFile sec.RSSFile, worklist []secworklist.Worklist) error {
	if s.Verbose {
		log.Info("Calculating number of XBRL Files in the index files: ")
	}

	totalCount, err := secutil.TotalXbrlFileCountGet(worklist, s, s.Config.Main.CacheDir)
	if err != nil {
		return err
	}
	if s.Verbose {
		log.Info(totalCount)
	}

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

	worklist, err := secworklist.WillDownloadGet(db)
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
					err = downloader.DownloadFile(db, v1.Enclosure.URL)
					if err != nil {
						return err
					}
					time.Sleep(rateLimit)
				}
			}
			currentCount++
			downloader.CurrentDownloadCount = currentCount
			if !s.Verbose {
				log.Info(fmt.Sprintf("\r[%d/%d/%f%% files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", currentCount, totalCount, downloader.GetDownloadPercentage(), (totalCount - currentCount)))
			}

			if s.Verbose {
				log.Info(fmt.Sprintf("[%d/%d/%f%%] %s downloaded...\n", currentCount, totalCount, downloader.GetDownloadPercentage(), time.Now().Format("2006-01-02 03:04:05")))
			}
		}

	}
	return nil
}

func DownloadFinancialStatementDataSets(db *sqlx.DB, s *sec.SEC) error {
	worklist, err := secworklist.WillDownloadGet(db)
	if err != nil {
		return err
	}

	downloader := download.NewDownloader(s.Config)
	downloader.IsContentLength = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = len(worklist)

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}
	for _, v := range worklist {
		quarter := secutil.QuarterFromMonth(v.Month)

		if !secutil.IsCurrentYearQuarterCorrect(v.Year, quarter) {
			continue
		}

		yearQuarter := fmt.Sprintf("%vq%v", v.Year, quarter)

		baseURL, err := url.Parse(s.BaseURL)
		if err != nil {
			return err
		}
		pathURL, err := url.Parse(fmt.Sprintf("/files/dera/data/financial-statement-data-sets/%v.zip", yearQuarter))
		if err != nil {
			return err
		}
		fileURL := baseURL.ResolveReference(pathURL).String()
		if s.Verbose {
			log.Info(fmt.Sprintf("Checking file '%v' in disk: ", filepath.Base(fileURL)))
		}
		isFileCorrect, err := downloader.FileCorrect(db, fileURL, 0, "")
		if err != nil {
			return err
		}
		if s.Verbose && isFileCorrect {
			log.Info("\u2713")
		}

		if !isFileCorrect {
			if s.Verbose {
				log.Info("Downloading file...: ")
			}
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			if s.Verbose {
				log.Info(time.Now().Format("2006-01-02 03:04:05"))
			}
			time.Sleep(rateLimit)
		}

		downloader.CurrentDownloadCount += 1
	}
	return nil
}

func DownloadRawFiles(s *sec.SEC, db *sqlx.DB) error {
	worklist, err := secworklist.WillDownloadGet(db)
	if err != nil {
		return err
	}

	type File struct {
		Path string
		Done bool
	}

	files := make(map[string]File)

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
			for _, v2 := range v1.XbrlFiling.XbrlFiles.XbrlFile {
				files[v2.URL] = File{
					Path: v2.URL,
					Done: false,
				}
			}
		}
	}

	var filesInDB []struct {
		XbrlURL      string
		XbrlFilePath string
	}

	err = db.Select(&filesInDB, "SELECT xbrlurl, xbrlfilepath FROM sec.secItemFile WHERE xbrlfilepath IS NOT NULL AND xbrlfilepath != '';")
	if err != nil {
		return err
	}

	for _, fn := range filesInDB {
		if file, ok := files[fn.XbrlURL]; ok {
			file.Done = true

			files[fn.XbrlURL] = file
		}
	}

	var filesToDownload []string
	for _, file := range files {
		if !file.Done {
			filesToDownload = append(filesToDownload, file.Path)
		}
	}

	if s.Verbose {
		log.Info("Number of files to be downloaded: ", len(filesToDownload))
	}

	downloader := download.NewDownloader(s.Config)
	downloader.IsEtag = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug
	downloader.TotalDownloadsCount = len(filesToDownload)
	downloader.CurrentDownloadCount = 1

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	for _, v := range filesToDownload {
		if s.Verbose {
			log.Info(fmt.Sprintf("Download progress [%d/%d/%f%%]", downloader.CurrentDownloadCount, downloader.TotalDownloadsCount, downloader.GetDownloadPercentage()))
		}

		err = downloader.DownloadFile(db, v)
		if err != nil {
			return err
		}
		time.Sleep(rateLimit)

		downloader.CurrentDownloadCount += 1
	}

	return nil
}
