// Copyright (c) 2021 Koszek Systems. All rights reserved.

package sec

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
)

const (
	XMLStartYear                           = 2005
	XMLStartMonth                          = 04
	FinancialStatementDataSetsStartYear    = 2009
	FinancialStatementDataSetsStartQuarter = 1
)

type RSSFile struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}
type Channel struct {
	Text          string `xml:",chardata"`
	Title         string `xml:"title"`
	Link          Link   `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	PubDate       string `xml:"pubDate"`
	LastBuildDate string `xml:"lastBuildDate"`
	Item          []Item `xml:"item"`
}

type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Atom string `xml:"atom,attr"`
}

type Item struct {
	Text        string     `xml:",chardata"`
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Guid        string     `xml:"guid"`
	Enclosure   Enclosure  `xml:"enclosure"`
	Description string     `xml:"description"`
	PubDate     string     `xml:"pubDate"`
	XbrlFiling  XbrlFiling `xml:"xbrlFiling"`
}

type Enclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type XbrlFiling struct {
	Text               string    `xml:",chardata"`
	Edgar              string    `xml:"edgar,attr"`
	CompanyName        string    `xml:"companyName"`
	FormType           string    `xml:"formType"`
	FilingDate         string    `xml:"filingDate"`
	CikNumber          string    `xml:"cikNumber"`
	AccessionNumber    string    `xml:"accessionNumber"`
	FileNumber         string    `xml:"fileNumber"`
	AcceptanceDatetime string    `xml:"acceptanceDatetime"`
	Period             string    `xml:"period"`
	AssistantDirector  string    `xml:"assistantDirector"`
	AssignedSic        string    `xml:"assignedSic"`
	FiscalYearEnd      string    `xml:"fiscalYearEnd"`
	XbrlFiles          XbrlFiles `xml:"xbrlFiles"`
}

type XbrlFiles struct {
	Text     string     `xml:",chardata"`
	XbrlFile []XbrlFile `xml:"xbrlFile"`
}

type XbrlFile struct {
	Text        string `xml:",chardata"`
	Sequence    string `xml:"sequence,attr"`
	File        string `xml:"file,attr"`
	Type        string `xml:"type,attr"`
	Size        string `xml:"size,attr"`
	Description string `xml:"description,attr"`
	InlineXBRL  string `xml:"inlineXBRL,attr"`
	URL         string `xml:"url,attr"`
}

type SECItemFile struct {
	ID                 int       `db:"id"`
	Ticker             string    `db:"ticker"`
	Title              string    `db:"title"`
	Link               string    `db:"link"`
	Guid               string    `db:"guid"`
	EnclosureURL       string    `db:"enclosure_url"`
	EnclosureLength    int       `db:"enclosure_length"`
	EnclosureType      string    `db:"enclosure_type"`
	Description        string    `db:"description"`
	PubDate            time.Time `db:"pubdate"`
	CompanyName        string    `db:"companyname"`
	FormType           string    `db:"formtype"`
	FillingDate        time.Time `db:"fillingdate"`
	CIKNumber          string    `db:"ciknumber"`
	AccessionNumber    string    `db:"accessionnumber"`
	FileNumber         string    `db:"filenumber"`
	AcceptanceDatetime string    `db:"acceptancedatetime"`
	Period             string    `db:"period"`
	AssistantDirector  string    `db:"assistantdirector"`
	AssignedSic        int       `db:"assignedsic"`
	FiscalYearEnd      int       `db:"fiscalyearend"`
	XbrlSequence       string    `db:"xbrlsequence"`
	XbrlFile           string    `db:"xbrlfile"`
	XbrlType           string    `db:"xbrltype"`
	XbrlSize           int       `db:"xbrlsize"`
	XbrlDescription    string    `db:"xbrldescription"`
	XbrlInlineXbrl     bool      `db:"xbrlinlinexbrl"`
	XbrlURL            string    `db:"xbrlurl"`
	XbrlBody           string    `db:"xbrlbody"`
}

type SEC struct {
	BaseURL string
	Verbose bool
	Config  config.Config
	Debug   bool
}

func NewSEC(config config.Config) (*SEC, error) {
	return &SEC{
		BaseURL: config.Main.BaseURL,
		Config:  config,
	}, nil
}

func (s *SEC) FetchFile(urlVar string) ([]byte, error) {
	// Retrieving JSON From URL
	baseURL, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, err
	}

	pathURL, err := url.Parse(urlVar)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(baseURL.ResolveReference(pathURL).String())
	if err != nil {
		return nil, err
	}

	// Reading JSON data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *SEC) ParseRSSGoXML(path string) (RSSFile, error) {
	var rssFile RSSFile

	xmlFile, err := os.Open(path)
	if err != nil {
		return rssFile, err
	}
	defer xmlFile.Close()

	data, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return rssFile, err
	}

	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&rssFile)
	if err != nil {
		return rssFile, err
	}

	return rssFile, err
}

func CalculateRSSFilesZIP(rssFile RSSFile) (int, error) {
	var totalSize int
	for _, v := range rssFile.Channel.Item {
		if v.Enclosure.Length != "" {
			val, err := strconv.Atoi(v.Enclosure.Length)
			if err != nil {
				return 0, err
			}
			totalSize += val
		}
	}
	return totalSize, nil
}

func ParseYearMonth(yearMonth string) (year int, month int, err error) {
	switch len(yearMonth) {
	case 4:
		date, err := time.Parse("2006", yearMonth)
		if err != nil {
			return 0, 0, err
		}
		year = date.Year()
	case 6:
		date, err := time.Parse("2006/1", yearMonth)
		if err != nil {
			return 0, 0, err
		}
		year = date.Year()
		month = int(date.Month())
	case 7:
		date, err := time.Parse("2006/01", yearMonth)
		if err != nil {
			return 0, 0, err
		}
		year = date.Year()
		month = int(date.Month())
	default:
		err := errors.New("please enter a valid date ('2021' or '2021/05')")
		return 0, 0, err
	}
	return year, month, nil
}

func (s *SEC) TotalXbrlFileCountGet(worklist []secworklist.Worklist, cacheDir string) (int, error) {
	var totalCount int
	for _, v := range worklist {
		filepath, err := s.FormatFilePathDate(cacheDir, v.Year, v.Month)
		if err != nil {
			return 0, err
		}

		rssFile, err := s.ParseRSSGoXML(filepath)
		if err != nil {
			return 0, err
		}

		for _, v1 := range rssFile.Channel.Item {
			totalCount += len(v1.XbrlFiling.XbrlFiles.XbrlFile)
		}
	}
	return totalCount, nil
}

func CheckRSSAvailability(year int, month int) (err error) {
	if year < XMLStartYear {
		log.Info(fmt.Sprintf("the earliest available XML is %d/%d", XMLStartYear, XMLStartMonth))
		return nil
	}

	if year == XMLStartYear && month > 0 && month < XMLStartMonth {
		log.Info(fmt.Sprintf("the earliest available XML is %d/%d", XMLStartYear, XMLStartMonth))
		return nil
	}

	if year > time.Now().Year() || month < 0 || month > 12 || (year == time.Now().Year() && month > int(time.Now().Month())) {
		log.Info(fmt.Sprintf("the latest available XML is %d/%d", time.Now().Year(), time.Now().Month()))
		return nil
	}

	return nil
}

func (s *SEC) DownloadToggle(db *sqlx.DB, year int, month int, willDownload bool) error {
	var err error

	for m := 1; m <= 12; m++ {
		if !IsMonthAvailable(year, m) {
			continue
		}

		if month != 0 && m != month {
			continue
		}

		err = secworklist.Save(year, m, willDownload, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func IsMonthAvailable(year int, month int) bool {
	if year > time.Now().Year() {
		return false
	}

	if year == time.Now().Year() && month > int(time.Now().Month()) {
		return false
	}

	if year < XMLStartYear {
		return false
	}

	if year == XMLStartMonth && month < XMLStartMonth {
		return false
	}

	return true
}

func (s *SEC) SearchByFilingDate(db *sqlx.DB, startdate time.Time, enddate time.Time) ([]SECItemFile, error) {
	secItemFiles := []SECItemFile{}
	err := db.Select(&secItemFiles, `
		SELECT sec.tickers.ticker, sec.secItemFile.title, sec.secItemFile.companyname, sec.secItemFile.ciknumber, sec.secItemFile. accessionnumber, sec.secItemFile.xbrlfile 
		FROM sec.secItemFile 
		LEFT JOIN sec.tickers
		ON sec.secitemfile.ciknumber = sec.tickers.cik
		WHERE DATE(fillingdate) between $1 AND $2
	`, startdate.Format("2006-01-02"), enddate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	return secItemFiles, nil
}

func (s *SEC) GetFilingDaysFromMonthYear(db *sqlx.DB, year int, month int) ([]int, error) {
	days := []int{}
	err := db.Select(&days, `
		SELECT DISTINCT EXTRACT(day from fillingdate)
		FROM sec.secItemFile 
		WHERE EXTRACT(year from fillingdate) = $1
		AND EXTRACT(month from fillingdate) = $2
	`, year, month)
	if err != nil {
		return nil, err
	}
	return days, nil
}

func (s *SEC) GetFilingCompaniesFromYearMonthDay(db *sqlx.DB, year int, month int, day int) ([]SECItemFile, error) {
	items := []SECItemFile{}
	err := db.Select(&items, `
		SELECT DISTINCT companyname, ciknumber
		FROM sec.secItemFile 
		WHERE EXTRACT(year from fillingdate) = $1
		AND EXTRACT(month from fillingdate) = $2
		AND EXTRACT(day from fillingdate) = $3
	`, year, month, day)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *SEC) SearchFilingsByYearMonthDayCIK(db *sqlx.DB, year int, month int, day int, cik int) ([]SECItemFile, error) {
	secItemFiles := []SECItemFile{}
	err := db.Select(&secItemFiles, `
		SELECT sec.tickers.ticker, sec.secItemFile.title, sec.secItemFile.companyname, sec.secItemFile.ciknumber, sec.secItemFile. accessionnumber, sec.secItemFile.xbrlfile 
		FROM sec.secItemFile 
		LEFT JOIN sec.tickers
		ON sec.secitemfile.ciknumber = sec.tickers.cik
		WHERE EXTRACT(year from fillingdate) = $1
		AND EXTRACT(month from fillingdate) = $2
		AND EXTRACT(day from fillingdate) = $3
		AND sec.secItemFile.cikNumber = $4;
	`, year, month, day, cik)
	if err != nil {
		return nil, err
	}
	return secItemFiles, nil
}

func (s *SEC) CreateFilesFromZIP(zipPath string, files []*zip.File) error {
	unpackedCachePath := filepath.Dir(filepath.Join(s.Config.Main.CacheDirUnpacked, zipPath))
	for _, file := range files {
		filePath := filepath.Join(unpackedCachePath, file.Name)

		isFileExists, err := os.Stat(filePath)
		if err != nil {
			if _, err = os.Stat(unpackedCachePath); err != nil {
				err = os.MkdirAll(unpackedCachePath, 0755)
				if err != nil {
					return err
				}
			}
		}

		if isFileExists == nil || (isFileExists != nil && isFileExists.Size() != file.FileInfo().Size()) {
			err = s.CreateFileFromZIP(file, filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SEC) CreateFileFromZIP(file *zip.File, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		log.Info("error in creating file for ZIP content file")
		return err
	}
	defer out.Close()

	reader, err := file.Open()
	if err != nil {
		log.Info("error in opening file from inside ZIP")
		return err
	}
	defer reader.Close()

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if err != nil {
		log.Info("error reading ZIP file content in buffer")
		return err
	}

	_, err = io.Copy(out, &buf)
	if err != nil {
		log.Info("error copying buffer content to file")
		return err
	}

	return nil
}

func (s *SEC) FormatFilePathDate(basepath string, year int, month int) (string, error) {
	date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", year, month))
	if err != nil {
		return "", err
	}
	formatted := date.Format("2006-01")

	filePath := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", basepath, formatted)
	return filePath, nil
}

func (s *SEC) ForEachWorklist(db *sqlx.DB, implementFunc func(*sqlx.DB, *SEC, RSSFile, []secworklist.Worklist) error, verboseMessage string) error {
	worklist, err := secworklist.WillDownloadGet(db)
	if err != nil {
		return err
	}
	for _, v := range worklist {
		fileURL, err := s.FormatFilePathDate(s.Config.Main.CacheDir, v.Year, v.Month)
		if err != nil {
			return err
		}

		_, err = os.Stat(fileURL)
		if err != nil {
			return fmt.Errorf("please run sec dow index to download all index files first")
		}

		rssFile, err := s.ParseRSSGoXML(fileURL)
		if err != nil {
			return err
		}

		if s.Verbose {
			log.Info(verboseMessage)
		}

		err = implementFunc(db, s, rssFile, worklist)
		if err != nil {
			return err
		}

	}
	return nil
}

func UnzipFiles(db *sqlx.DB, s *SEC, rssFile RSSFile, worklist []secworklist.Worklist) error {
	totalCount := len(rssFile.Channel.Item)
	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		parsedURL, err := url.Parse(v1.Enclosure.URL)
		if err != nil {
			return err
		}
		zipPath := parsedURL.Path

		zipCachePath := filepath.Join(s.Config.Main.CacheDir, zipPath)
		_, err = os.Stat(zipCachePath)
		if err != nil {
			log.Info("please run sec dowz to download all ZIP files then run sec indexz again to index them")
			return err
		}

		if strings.ToLower(filepath.Ext(zipCachePath)) != ".zip" {
			continue
		}

		reader, err := zip.OpenReader(zipCachePath)
		if err != nil {
			log.Info("error opening the file:", zipCachePath)
			return err
		}

		err = s.CreateFilesFromZIP(zipPath, reader.File)
		if err != nil {
			log.Info("error creating files from ZIP:", zipPath)
			return err
		}

		reader.Close()

		currentCount++
		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d] %s unpacked...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
		}
	}
	return nil
}

func QuarterFromMonth(month int) int {
	if month >= 1 && month <= 3 {
		return 1
	}
	if month >= 4 && month <= 6 {
		return 2
	}
	if month >= 7 && month <= 9 {
		return 3
	}
	if month >= 10 && month <= 12 {
		return 4
	}
	return 0
}

func IsCurrentYearQuarterCorrect(year int, quarter int) bool {
	// Below is to check if quarter is not in sec.gov website
	// For example: If we are in August (8) then we can only download up to Q2 (June)
	// because we did not complete Q3 (We complete it after we END Sept (9))
	var currentMonth = int(time.Now().Month())
	if year == time.Now().Year() && ((quarter == 1 && currentMonth < 4) ||
		(quarter == 2 && currentMonth < 7) ||
		(quarter == 3 && currentMonth < 10) ||
		(quarter == 4 && currentMonth < 12)) {
		return false
	}
	return true
}

func GetFailedDownloadEventCount(db *sqlx.DB) (int, error) {
	var count []int
	err := db.Select(&count, "SELECT COUNT(*) FROM sec.events WHERE ev ->> 'event' = 'download' AND ev ->> 'status' = 'failed'")
	if err != nil {
		return 0, err
	}

	if len(count) > 0 {
		return count[0], nil
	}

	return 0, nil
}

func GetSuccessfulDownloadEventCount(db *sqlx.DB) (int, error) {
	var count []int
	err := db.Select(&count, "SELECT COUNT(*) FROM sec.events WHERE ev ->> 'event' = 'download' AND ev ->> 'status' = 'success'")
	if err != nil {
		return 0, err
	}

	if len(count) > 0 {
		return count[0], nil
	}

	return 0, nil
}

func (s SEC) GetTotalZIPFilesToBeDownloaded(db *sqlx.DB, worklist []secworklist.Worklist) (int, error) {
	if s.Verbose {
		log.Info("Getting Number of ZIP Files To Be Downloaded...")
	}

	var totalZIPFilesToBeDownloaded int
	for _, v := range worklist {
		fileURL, err := s.FormatFilePathDate(s.Config.Main.CacheDir, v.Year, v.Month)
		if err != nil {
			return 0, err
		}

		_, err = os.Stat(fileURL)
		if err != nil {
			return 0, fmt.Errorf("please run sec dow index to download all index files first")
		}

		rssFile, err := s.ParseRSSGoXML(fileURL)
		if err != nil {
			return 0, err
		}

		totalZIPFilesToBeDownloaded += len(rssFile.Channel.Item)
	}

	if s.Verbose {
		log.Info("There is a total of ", totalZIPFilesToBeDownloaded, " ZIP files to be downloaded.")
	}
	return totalZIPFilesToBeDownloaded, nil
}

func GetFiveRecentFilings(db *sqlx.DB) ([]SECItemFile, error) {
	var secitemfiles []SECItemFile

	err := db.Select(&secitemfiles, `
	SELECT companyname, formtype, pubdate, xbrlurl 
	FROM sec.secitemfile
	WHERE RIGHT(xbrlurl, 3) = 'htm' 
		OR RIGHT(xbrlurl, 4) = 'html'
		OR RIGHT(xbrlurl, 3) = 'xml'
	ORDER BY created_at desc LIMIT 5;`)
	if err != nil {
		return nil, err
	}

	return secitemfiles, nil
}
