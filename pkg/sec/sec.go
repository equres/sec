// Copyright (c) 2021 Koszek Systems. All rights reserved.

package sec

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
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
	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
	"jaytaylor.com/html2text"
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

type ExchangesFile struct {
	Fields []string        `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

type Worklist struct {
	Year         int  `db:"year"`
	Month        int  `db:"month"`
	WillDownload bool `db:"will_download"`
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

type SUB struct {
	Adsh       string `csv:"adsh"`
	CIK        int    `csv:"cik"`
	Name       string `csv:"name"`
	SIC        string `csv:"sic"`
	CountryBA  string `csv:"countryba"`
	StprBA     string `csv:"stprba"`
	CityBA     string `csv:"cityba"`
	ZIPBA      string `csv:"zipba"`
	BAS1       string `csv:"bas1"`
	BAS2       string `csv:"bas2"`
	BAPH       string `csv:"baph"`
	CountryMA  string `csv:"countryma"`
	StrpMA     string `csv:"strpma"`
	CityMA     string `csv:"cityma"`
	ZIPMA      string `csv:"zipma"`
	MAS1       string `csv:"mas1"`
	MAS2       string `csv:"mas2"`
	CountryInc string `csv:"countryinc"`
	StprInc    string `csv:"stprinc"`
	EIN        string `csv:"ein"`
	Former     string `csv:"former"`
	Changed    string `csv:"changed"`
	Afs        string `csv:"afs"`
	Wksi       string `csv:"wksi"`
	Fye        string `csv:"fye"`
	Form       string `csv:"form"`
	Period     string `csv:"period"`
	Fy         string `csv:"fy"`
	Fp         string `csv:"fp"`
	Filled     string `csv:"filled"`
	Accepted   string `csv:"accepted"`
	Prevrpt    string `csv:"prevrpt"`
	Detail     string `csv:"detail"`
	Instance   string `csv:"instance"`
	Nciks      string `csv:"nciks"`
	Aciks      string `csv:"aciks"`
}

type NUM struct {
	Adsh     string `csv:"adsh"`
	Tag      string `csv:"tag"`
	Version  string `csv:"version"`
	Coreg    string `csv:"coreg"`
	DDate    string `csv:"ddate"`
	Qtrs     string `csv:"qtrs"`
	UOM      string `csv:"uom"`
	Value    string `csv:"value"`
	Footnote string `csv:"footnote"`
}

type TAG struct {
	Tag      string `csv:"tag"`
	Version  string `csv:"version"`
	Custom   string `csv:"custom"`
	Abstract string `csv:"abstract"`
	Datatype string `csv:"datatype"`
	Lord     string `csv:"lord"`
	Crdr     string `csv:"crdr"`
	Tlabel   string `csv:"tlabel"`
	Doc      string `csv:"doc"`
}

type PRE struct {
	Adsh    string `csv:"adsh"`
	Report  string `csv:"report"`
	Line    string `csv:"line"`
	Stmt    string `csv:"stmt"`
	Inpth   string `csv:"inpth"`
	Rfile   string `csv:"rfile"`
	Tag     string `csv:"tag"`
	Version string `csv:"version"`
	Plabel  string `csv:"plabel"`
}

// Ticker Struct Based on JSON
type SecTicker struct {
	Cik      int    `json:"cik_str"`
	Ticker   string `json:"ticker"`
	Title    string `json:"title"`
	Exchange string `json:"exchange"`
}

type SEC struct {
	BaseURL string
	Tickers []SecTicker
	Verbose bool
	Config  config.Config
	Debug   bool
}

func (t SecTicker) String() string {
	return fmt.Sprintf("Cik: %d\nTicker: %s\nTitle: %s\nExchange: %s\n", t.Cik, t.Ticker, t.Title, t.Exchange)
}

func NewSEC(config config.Config) (*SEC, error) {
	return &SEC{
		BaseURL: config.Main.BaseURL,
		Config:  config,
	}, nil
}

func (t SecTicker) Save(db *sqlx.DB) error {
	_, err := db.Exec(`
		INSERT INTO sec.tickers (cik, ticker, title, exchange, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, NOW(), NOW()) 
		ON CONFLICT (cik, ticker, title) 
		DO UPDATE SET
			cik=EXCLUDED.cik,
			ticker=EXCLUDED.ticker,
			title=EXCLUDED.title,
			exchange=EXCLUDED.exchange,
			updated_at=NOW() 
		WHERE 1=1
		AND tickers.cik=EXCLUDED.cik
		AND tickers.ticker=EXCLUDED.ticker
		AND tickers.title=EXCLUDED.title
		AND tickers.exchange = '';`, t.Cik, t.Ticker, t.Title, t.Exchange)
	if err != nil {
		return err
	}
	return nil
}

func SaveCIK(db *sqlx.DB, cik int) error {
	_, err := db.Exec(`
		INSERT INTO sec.ciks (cik, created_at, updated_at) 
		VALUES ($1,NOW(), NOW()) 
		ON CONFLICT (cik) 
		DO NOTHING;`, cik)
	if err != nil {
		return err
	}
	return nil
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

func (s *SEC) TickerUpdateAll(db *sqlx.DB) error {
	err := s.NoExchangeTickersGet(db)
	if err != nil {
		return err
	}

	err = s.ExchangeTickersGet(db)
	if err != nil {
		return err
	}
	return nil
}

func (s *SEC) DownloadTickerFile(db *sqlx.DB, path string) error {
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

func (s *SEC) TickersGetAll(db *sqlx.DB) ([]SecTicker, error) {
	// Retrieve from DB
	tickers := []SecTicker{}
	err := db.Select(&tickers, "SELECT cik, ticker, title, exchange FROM sec.tickers")
	if err != nil {
		return nil, err
	}
	return tickers, nil
}

func (s *SEC) NoExchangeTickersGet(db *sqlx.DB) error {
	file, err := os.Open(filepath.Join(s.Config.Main.CacheDir, "files/company_tickers.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// Creating Map to hold company ticker structs
	allCompanyTickers := make(map[int]SecTicker)

	// Converting JSON to Structs
	err = json.Unmarshal(data, &allCompanyTickers)
	if err != nil {
		return err
	}

	if s.Verbose {
		log.Info("Indexing file company_tickers.json: ")
	}

	for _, v := range allCompanyTickers {
		err = SaveCIK(db, v.Cik)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers.json", "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	for _, v := range allCompanyTickers {
		sec := SecTicker{
			Cik:      v.Cik,
			Ticker:   v.Ticker,
			Title:    v.Title,
			Exchange: v.Exchange,
		}
		err := sec.Save(db)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers.json", "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	if s.Verbose {
		log.Info("\u2713")
	}
	eventErr := database.CreateIndexEvent(db, "company_tickers.json", "success")
	if eventErr != nil {
		return eventErr
	}

	return nil
}

func (s *SEC) ExchangeTickersGet(db *sqlx.DB) error {
	// Retrieving JSON data from URL
	file, err := os.Open(filepath.Join(s.Config.Main.CacheDir, "files/company_tickers_exchange.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	fileExchange := ExchangesFile{}

	err = json.Unmarshal(data, &fileExchange)
	if err != nil {
		return err
	}

	if s.Verbose {
		log.Info("Indexing file company_tickers_exchange.json: ")
	}

	for _, v := range fileExchange.Data {
		err = SaveCIK(db, int(v[0].(float64)))
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	for _, v := range fileExchange.Data {
		// Below is because sometimes the exchange is empty (nil). Added lines to ensure no error when saving
		exchange := ""
		if v[3] == nil {
			exchange = ""
		} else {
			exchange = v[3].(string)
		}

		sec := SecTicker{
			Cik:      int(v[0].(float64)),
			Title:    v[1].(string),
			Ticker:   v[2].(string),
			Exchange: exchange,
		}
		err := sec.Save(db)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}
	if s.Verbose {
		log.Info("\u2713")
	}
	eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "success")
	if eventErr != nil {
		return eventErr
	}
	return nil
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

func (s *SEC) DownloadIndex(db *sqlx.DB) error {
	worklist, err := WorklistWillDownloadGet(db)
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
		fileURL, err := s.FormatFilePathDate(s.BaseURL, v.Year, v.Month)
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

func (s *SEC) CalculateRSSFilesZIP(rssFile RSSFile) (int, error) {
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

func SaveWorklist(year int, month int, willDownload bool, db *sqlx.DB) error {
	_, err := db.Exec(`
		INSERT INTO sec.worklist (year, month, will_download, created_at, updated_at) 
		VALUES ($1, $2, $3, NOW(), NOW()) 
		ON CONFLICT (month, year) 
		DO UPDATE SET
			will_download = EXCLUDED.will_download,
			updated_at=NOW() 
		WHERE 1=1
		AND worklist.year=EXCLUDED.year
		AND worklist.month=EXCLUDED.month ;`, year, month, willDownload)
	if err != nil {
		return err
	}
	return nil
}

func (s *SEC) SecItemFileUpsert(db *sqlx.DB, item Item) error {
	var err error

	var enclosureLength int
	if item.Enclosure.Length != "" {
		enclosureLength, err = strconv.Atoi(item.Enclosure.Length)
		if err != nil {
			return err
		}
	}
	var assignedSic int
	if item.XbrlFiling.AssignedSic != "" {
		assignedSic, err = strconv.Atoi(item.XbrlFiling.AssignedSic)
		if err != nil {
			return err
		}
	}

	var fiscalYearEnd int
	if item.XbrlFiling.FiscalYearEnd != "" {
		fiscalYearEnd, err = strconv.Atoi(item.XbrlFiling.FiscalYearEnd)
		if err != nil {
			return err
		}
	}
	var cikNumber int
	if item.XbrlFiling.CikNumber != "" {
		cikNumber, err = strconv.Atoi(item.XbrlFiling.CikNumber)
		if err != nil {
			return err
		}
	}

	// Check if CIK here is in CIKs table
	var ciks []int
	err = db.Select(&ciks, "SELECT cik FROM sec.ciks WHERE cik = $1", cikNumber)
	if err != nil {
		return err
	}
	if len(ciks) == 0 {
		return nil
	}

	for _, v := range item.XbrlFiling.XbrlFiles.XbrlFile {
		var xbrlInline bool
		if v.InlineXBRL != "" {
			xbrlInline, err = strconv.ParseBool(v.InlineXBRL)
			if err != nil {
				return err
			}
		}

		var xbrlSequence int
		if v.Sequence != "" {
			xbrlSequence, err = strconv.Atoi(v.Sequence)
			if err != nil {
				return err
			}
		}

		var xbrlSize int
		if v.Size != "" {
			xbrlSize, err = strconv.Atoi(v.Size)
			if err != nil {
				return err
			}
		}

		fileUrl, err := url.Parse(v.URL)
		if err != nil {
			return err
		}

		var fileBody string
		filePath := filepath.Join(s.Config.Main.CacheDir, fileUrl.Path)
		_, err = os.Stat(filePath)
		if err == nil {
			xbrlFile, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer xbrlFile.Close()

			data, err := ioutil.ReadAll(xbrlFile)
			if err != nil {
				return err
			}

			if s.IsFileIndexable(filePath) {
				fileBody = string(data)
				if s.IsFileTypeHTML(filePath) {
					fileBody, err = html2text.FromString(string(data))
					if err != nil {
						return err
					}
				}
			}
		} else {
			log.Errorf("Could not find/index the file %v", filePath)
		}

		_, err = db.Exec(`
		INSERT INTO sec.secItemFile (title, link, guid, enclosure_url, enclosure_length, enclosure_type, description, pubdate, companyname, formtype, fillingdate, ciknumber, accessionnumber, filenumber, acceptancedatetime, period, assistantdirector, assignedsic, fiscalyearend, xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl, xbrlbody, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, NOW(), NOW()) 

		ON CONFLICT (xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl)
		DO UPDATE SET title=EXCLUDED.title, link=EXCLUDED.link, guid=EXCLUDED.guid, enclosure_url=EXCLUDED.enclosure_url, enclosure_length=EXCLUDED.enclosure_length, enclosure_type=EXCLUDED.enclosure_type, description=EXCLUDED.description, pubdate=EXCLUDED.pubdate, companyname=EXCLUDED.companyname, formtype=EXCLUDED.formtype, fillingdate=EXCLUDED.fillingdate, ciknumber=EXCLUDED.ciknumber, accessionnumber=EXCLUDED.accessionnumber, filenumber=EXCLUDED.filenumber, acceptancedatetime=EXCLUDED.acceptancedatetime, period=EXCLUDED.period, assistantdirector=EXCLUDED.assistantdirector, assignedsic=EXCLUDED.assignedsic, fiscalyearend=EXCLUDED.fiscalyearend, xbrlsequence=EXCLUDED.xbrlsequence, xbrlfile=EXCLUDED.xbrlfile, xbrltype=EXCLUDED.xbrltype, xbrlsize=EXCLUDED.xbrlsize, xbrldescription=EXCLUDED.xbrldescription, xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl, xbrlurl=EXCLUDED.xbrlurl, updated_at=NOW()
		WHERE secItemFile.xbrlsequence=EXCLUDED.xbrlsequence AND secItemFile.xbrlfile=EXCLUDED.xbrlfile AND secItemFile.xbrltype=EXCLUDED.xbrltype AND secItemFile.xbrlsize=EXCLUDED.xbrlsize AND secItemFile.xbrldescription=EXCLUDED.xbrldescription AND secItemFile.xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl AND secItemFile.xbrlurl=EXCLUDED.xbrlurl AND secItemFile.xbrlbody=EXCLUDED.xbrlbody;`,
			item.Title, item.Link, item.Guid, item.Enclosure.URL, enclosureLength, item.Enclosure.Type, item.Description, item.PubDate, item.XbrlFiling.CompanyName, item.XbrlFiling.FormType, item.XbrlFiling.FilingDate, cikNumber, item.XbrlFiling.AccessionNumber, item.XbrlFiling.FileNumber, item.XbrlFiling.AcceptanceDatetime, item.XbrlFiling.Period, item.XbrlFiling.AssistantDirector, assignedSic, fiscalYearEnd, xbrlSequence, v.File, v.Type, xbrlSize, v.Description, xbrlInline, v.URL, fileBody)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, filePath, "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}

		eventErr := database.CreateIndexEvent(db, filePath, "success")
		if eventErr != nil {
			return eventErr
		}
	}
	return nil
}

func (s *SEC) IsFileIndexable(filename string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filename))

	if fileExtension == ".html" || fileExtension == ".htm" || fileExtension == ".xml" {
		return true
	}
	return false
}

func (s *SEC) IsFileTypeHTML(filename string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filename))

	if fileExtension == ".html" || fileExtension == ".htm" {
		return true
	}
	return false
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

func (s *SEC) TotalXbrlFileCountGet(worklist []Worklist, cacheDir string) (int, error) {
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

func (s *SEC) DownloadXbrlFileContent(db *sqlx.DB, files []XbrlFile, config config.Config, currentCount *int, totalCount int) error {
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
			log.Info(fmt.Sprintf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", *currentCount, totalCount, (totalCount - *currentCount)))
		}

		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d] %s downloaded...\n", *currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
		}
		time.Sleep(rateLimit)
	}

	return nil
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

func (s *SEC) Downloadability(db *sqlx.DB, year int, month int, willDownload bool) error {
	var err error

	if month != 0 {
		err = SaveWorklist(year, month, willDownload, db)
		if err != nil {
			return err
		}
		return nil
	}

	firstMonthAvailable := XMLStartMonth
	if year > XMLStartYear {
		firstMonthAvailable = 1
	}

	lastMonthAvailable := 12
	if year == time.Now().Year() {
		lastMonthAvailable = int(time.Now().Month())
	}

	for i := firstMonthAvailable; i <= lastMonthAvailable; i++ {
		err = SaveWorklist(year, i, willDownload, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func WorklistWillDownloadGet(db *sqlx.DB) ([]Worklist, error) {
	// Retrieve from DB
	var worklist []Worklist

	err := db.Select(&worklist, "SELECT year, month, will_download FROM sec.worklist WHERE will_download = true ORDER BY year, month ASC")
	if err != nil {
		return nil, err
	}
	return worklist, nil
}

func UniqueYearsInWorklist(db *sqlx.DB) ([]int, error) {
	// Retrieve from DB
	var worklistYears []int

	err := db.Select(&worklistYears, "SELECT DISTINCT year FROM sec.worklist WHERE will_download = true ORDER BY year ASC")
	if err != nil {
		return nil, err
	}
	return worklistYears, nil
}

func MonthsInYearInWorklist(db *sqlx.DB, year int) ([]int, error) {
	// Retrieve from DB
	var worklistMonths []int

	err := db.Select(&worklistMonths, "SELECT month FROM sec.worklist WHERE will_download = true AND year = $1 ORDER BY year ASC", year)
	if err != nil {
		return nil, err
	}
	return worklistMonths, nil
}

func (s *SEC) ZIPContentUpsert(db *sqlx.DB, pathname string, files []*zip.File) error {
	// Keeping only directories
	dirsPath := filepath.Dir(pathname)

	// Spliting directories
	dirs := strings.Split(dirsPath, "\\")
	if len(dirs) == 1 {
		dirs = strings.Split(dirsPath, "/")
	}

	// Keeping only CIK and Accession Number
	dirs = dirs[len(dirs)-2:]

	cik := dirs[0]
	accession := dirs[1]

	// Check if CIK here is in CIKs table
	var ciks []int
	err := db.Select(&ciks, "SELECT cik FROM sec.ciks WHERE cik = $1", cik)
	if err != nil {
		return err
	}
	if len(ciks) == 0 {
		return nil
	}

	for _, file := range files {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(reader)
		if err != nil {
			return err
		}
		var xbrlBody string

		if s.IsFileIndexable(file.FileInfo().Name()) {
			xbrlBody = buf.String()
		}

		_, err = db.Exec(`
			INSERT INTO sec.secItemFile (ciknumber, accessionnumber, xbrlfile, xbrlsize, xbrlbody, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) 
			ON CONFLICT (cikNumber, accessionNumber, xbrlFile, xbrlSize)
			DO NOTHING;`, cik, accession, file.Name, int(file.FileInfo().Size()), xbrlBody)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, pathname, "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	eventErr := database.CreateIndexEvent(db, pathname, "success")
	if eventErr != nil {
		return eventErr
	}
	return nil
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

func (s *SEC) DownloadAllItemFiles(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
	if s.Verbose {
		log.Info("Calculating number of XBRL Files in the index files: ")
	}

	totalCount, err := s.TotalXbrlFileCountGet(worklist, s.Config.Main.CacheDir)
	if err != nil {
		return err
	}
	if s.Verbose {
		log.Info(totalCount)
	}

	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		err := s.DownloadXbrlFileContent(db, v1.XbrlFiling.XbrlFiles.XbrlFile, s.Config, &currentCount, totalCount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SEC) ForEachWorklist(db *sqlx.DB, implementFunc func(*sqlx.DB, RSSFile, []Worklist) error, verboseMessage string) error {
	worklist, err := WorklistWillDownloadGet(db)
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

		err = implementFunc(db, rssFile, worklist)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *SEC) DownloadZIPFiles(db *sqlx.DB) error {
	downloader := download.NewDownloader(s.Config)
	downloader.IsContentLength = true
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	worklist, err := WorklistWillDownloadGet(db)
	if err != nil {
		return err
	}

	totalCount, err := s.GetTotalZIPFilesToBeDownloaded(db, worklist)
	if err != nil {
		return err
	}
	currentCount := 0

	downloader.CurrentDownloadCount = 0
	downloader.TotalDownloadsCount = totalCount
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

				currentCount++
				downloader.CurrentDownloadCount = currentCount
				if !s.Verbose {
					log.Info(fmt.Sprintf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", currentCount, totalCount, (totalCount - currentCount)))
				}

				if s.Verbose {
					log.Info(fmt.Sprintf("[%d/%d] %s downloaded...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
				}
			}
		}

	}
	return nil
}

func (s *SEC) IndexZIPFileContent(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
	totalCount := len(rssFile.Channel.Item)
	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		if v1.Enclosure.URL == "" {
			continue
		}
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

		reader, err := zip.OpenReader(zipCachePath)
		if err != nil {
			log.Errorf("Could not access the file %v", zipCachePath)
			continue
		}

		err = s.ZIPContentUpsert(db, zipPath, reader.File)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, zipCachePath, "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}

		eventErr := database.CreateIndexEvent(db, zipCachePath, "success")
		if eventErr != nil {
			return eventErr
		}

		reader.Close()
		currentCount++

		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d] %s inserted for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
		}
	}
	return nil
}

func (s *SEC) UnzipFiles(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
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

func (s *SEC) InsertAllSecItemFile(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
	totalCount := len(rssFile.Channel.Item)
	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		err := s.SecItemFileUpsert(db, v1)
		if err != nil {
			return err
		}
		currentCount++
		if s.Verbose {
			log.Info(fmt.Sprintf("[%d/%d] %s inserted for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05")))
		}
	}
	return nil
}

func (s *SEC) DownloadFinancialStatementDataSets(db *sqlx.DB) error {
	worklist, err := WorklistWillDownloadGet(db)
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
		quarter := QuarterFromMonth(v.Month)

		if !IsCurrentYearQuarterCorrect(v.Year, quarter) {
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

func (s *SEC) IndexFinancialStatementDataSets(db *sqlx.DB) error {
	filesPath := filepath.Join(s.Config.Main.CacheDir, "files/dera/data/financial-statement-data-sets/")
	files, err := ioutil.ReadDir(filesPath)
	if err != nil {
		return err
	}
	for _, v := range files {
		if s.Verbose {
			log.Info(fmt.Sprintf("Indexing file %v: ", v.Name()))
		}
		reader, err := zip.OpenReader(filepath.Join(filesPath, v.Name()))
		if err != nil {
			return err
		}
		defer reader.Close()

		err = s.FinancialStatementDataSetsZIPUpsert(db, filesPath, reader.File)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, filesPath, "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
		if s.Verbose {
			log.Info("\u2713")
		}
		eventErr := database.CreateIndexEvent(db, filesPath, "success")
		if eventErr != nil {
			return eventErr
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

func (s *SEC) FinancialStatementDataSetsZIPUpsert(db *sqlx.DB, pathname string, files []*zip.File) error {
	for _, file := range files {
		fileName := strings.ToLower(file.Name)

		if fileName == "readme.htm" || fileName == "readme.html" {
			continue
		}

		var upsertFunc func(*sqlx.DB, io.ReadCloser) error
		switch fileName {
		case "sub.txt":
			upsertFunc = s.SubDataUpsert
		case "tag.txt":
			upsertFunc = s.TagDataUpsert
		case "num.txt":
			upsertFunc = s.NumDataUpsert
		case "pre.txt":
			upsertFunc = s.PreDataUpsert
		default:
			continue
		}

		if s.Verbose {
			log.Info(fmt.Sprintf("Indexing file %v\n", fileName))
		}

		reader, err := file.Open()
		if err != nil {
			eventErr := database.CreateIndexEvent(db, pathname, "failed")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
		defer reader.Close()

		err = upsertFunc(db, reader)
		if err != nil {
			return err
		}
	}
	eventErr := database.CreateIndexEvent(db, pathname, "failed")
	if eventErr != nil {
		return eventErr
	}
	return nil
}

func (s *SEC) SubDataUpsert(db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	subs := []SUB{}
	err = gocsv.Unmarshal(reader, &subs)
	if err != nil {
		return err
	}

	for _, v := range subs {
		ciks := []struct {
			CIK string
		}{}
		err = db.Select(&ciks, "SELECT cik FROM sec.ciks WHERE cik = $1", v.CIK)
		if err != nil {
			return err
		}

		if len(ciks) == 0 {
			continue
		}

		var period sql.NullString
		if v.Period != "" {
			period = sql.NullString{
				String: v.Period,
			}
		}

		var filled sql.NullString
		if v.Filled != "" {
			filled = sql.NullString{
				String: v.Filled,
			}
		}

		_, err = db.Exec(`
		INSERT INTO sec.sub (adsh, cik, name, sic, countryba, stprba, cityba, zipba, bas1, bas2, baph, countryma, strpma, cityma, zipma, mas1, mas2, countryinc, stprinc, ein, former, changed, afs, wksi, fye, form, period, fy, fp, filled, accepted, prevrpt, detail, instance, nciks, aciks, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, NOW(), NOW()) 
		ON CONFLICT (adsh, cik, name, sic) 
		DO NOTHING;`, v.Adsh, v.CIK, v.Name, v.SIC, v.CountryBA, v.StprBA, v.CityBA, v.ZIPBA, v.BAS1, v.BAS2, v.BAPH, v.CountryMA, v.StrpMA, v.CityMA, v.ZIPMA, v.MAS1, v.MAS2, v.CountryInc, v.StprInc, v.EIN, v.Former, v.Changed, v.Afs, v.Wksi, v.Fye, v.Form, period, v.Fy, v.Fp, filled, v.Accepted, v.Prevrpt, v.Detail, v.Instance, v.Nciks, v.Aciks)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SEC) TagDataUpsert(db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	tags := []TAG{}
	err = gocsv.Unmarshal(reader, &tags)
	if err != nil {
		return err
	}

	for _, v := range tags {
		_, err = db.Exec(`
			INSERT INTO sec.tag (tag, version, custom, abstract, datatype, lord, crdr, tlabel, doc, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (tag, version) 
			DO NOTHING;`, v.Tag, v.Version, v.Custom, v.Abstract, v.Datatype, v.Lord, v.Crdr, v.Tlabel, v.Doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SEC) NumDataUpsert(db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	nums := []NUM{}
	err = gocsv.Unmarshal(reader, &nums)
	if err != nil {
		return err
	}

	for _, v := range nums {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM sec.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO sec.num (adsh, tag, version, coreg, ddate, qtrs, uom, value, footnote, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, tag, version, coreg, ddate, qtrs, uom) 
			DO NOTHING;`, v.Adsh, v.Tag, v.Version, v.Coreg, v.DDate, v.Qtrs, v.UOM, v.Value, v.Footnote)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SEC) PreDataUpsert(db *sqlx.DB, reader io.ReadCloser) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		return r
	})

	pres := []PRE{}
	err = gocsv.Unmarshal(reader, &pres)
	if err != nil {
		return err
	}

	for _, v := range pres {
		tags := []struct {
			Tag     string `db:"tag"`
			Version string `db:"version"`
		}{}
		err = db.Select(&tags, "SELECT tag, version FROM sec.tag WHERE tag = $1 AND version = $2;", v.Tag, v.Version)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO sec.pre (adsh, report, line, stmt, inpth, rfile, tag, version, plabel, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) 
			ON CONFLICT (adsh, report, line) 
			DO NOTHING;`, v.Adsh, v.Report, v.Line, v.Stmt, v.Inpth, v.Rfile, v.Tag, v.Version, v.Plabel)
		if err != nil {
			return err
		}
	}
	return nil
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

func (s SEC) GetTotalZIPFilesToBeDownloaded(db *sqlx.DB, worklist []Worklist) (int, error) {
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
