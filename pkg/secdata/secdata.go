package secdata

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
)

type SECDataOps interface {
	GetDataType() string
	GetDataFilePath(baseURL string, yearQuarter string) (string, error)
	GetDataDirPath() string
	GetDataTypeInsertFunc(fileName string) func(*sec.SEC, *sqlx.DB, io.ReadCloser) error
}

type SECData struct {
	SECDataOps SECDataOps
}

func NewSECData(s SECDataOps) *SECData {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})
	return &SECData{
		SECDataOps: s,
	}
}

func (sd *SECData) DownloadSECData(db *sqlx.DB, s *sec.SEC) error {
	worklist, err := secworklist.WillDownloadGet(db, true)
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

		fileURL, err := sd.SECDataOps.GetDataFilePath(s.BaseURL, yearQuarter)
		if err != nil {
			return err
		}

		s.Log(fmt.Sprintf("Checking file '%v' in disk: ", filepath.Base(fileURL)))
		isFileCorrect, err := downloader.FileCorrect(db, fileURL, 0, "")
		if err != nil {
			return err
		}
		if isFileCorrect {
			s.Log("\u2713")
		}

		if !isFileCorrect {
			s.Log("Downloading file...: ")
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			s.Log(time.Now().Format("2006-01-02 03:04:05"))
			time.Sleep(rateLimit)
		}

		downloader.CurrentDownloadCount += 1
	}
	return nil
}

func (sd *SECData) IndexData(s *sec.SEC, db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, sd.SECDataOps.GetDataDirPath())
	files, err := ioutil.ReadDir(filesPath)
	if err != nil {
		return err
	}
	for _, v := range files {
		s.Log(fmt.Sprintf("Indexing file %v: ", v.Name()))
		reader, err := zip.OpenReader(filepath.Join(filesPath, v.Name()))
		if err != nil {
			return err
		}

		err = sd.ZIPFileUpsert(s, db, filesPath, reader.File)
		if err != nil {
			secevent.CreateIndexEvent(db, filesPath, "failed", "error_inserting_secdata_in_database")
			return err
		}

		reader.Close()

		s.Log("\u2713")
		secevent.CreateIndexEvent(db, filesPath, "success", "")
	}
	return nil
}

func (sd *SECData) ZIPFileUpsert(s *sec.SEC, db *sqlx.DB, pathname string, files []*zip.File) error {
	for _, file := range files {
		fileName := strings.ToLower(file.Name)

		if fileName == "readme.htm" || fileName == "readme.html" || fileName == "rr1-metadata.json" {
			continue
		}

		upsertFunc := sd.SECDataOps.GetDataTypeInsertFunc(fileName)
		if upsertFunc == nil {
			return fmt.Errorf("could_not_identify_file_type_func %v", fileName)
		}

		s.Log(fmt.Sprintf("Indexing file %v\n", fileName))

		reader, err := file.Open()
		if err != nil {
			secevent.CreateIndexEvent(db, pathname, "failed", "could_not_open_zip_file")
			return err
		}

		err = upsertFunc(s, db, reader)
		if err != nil {
			return err
		}

		reader.Close()
	}
	secevent.CreateIndexEvent(db, pathname, "success", "")
	return nil
}
