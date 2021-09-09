// Copyright (c) 2021 Equres LLC. All rights reserved.

package sec

import (
	"archive/zip"
	"bytes"
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

	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/download"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
	"jaytaylor.com/html2text"
)

const (
	XMLStartYear  = 2005
	XMLStartMonth = 04
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
	DO UPDATE SET cik=EXCLUDED.cik, ticker=EXCLUDED.ticker, title=EXCLUDED.title, exchange=EXCLUDED.exchange, updated_at=NOW() 
	WHERE tickers.cik=EXCLUDED.cik AND tickers.ticker=EXCLUDED.ticker AND tickers.title=EXCLUDED.title AND tickers.exchange = '';`, t.Cik, t.Ticker, t.Title, t.Exchange)
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
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

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
		fmt.Printf("Checking for file %v: ", filepath.Base(pathURL.Path))
	}

	isFileCorrect, err := downloader.FileCorrect(db, fullURL)
	if err != nil {
		return err
	}

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	if s.Verbose && isFileCorrect {
		fmt.Println("\u2713")
	}
	if !isFileCorrect {
		if s.Verbose {
			fmt.Print("Downloading file...: ")
		}
		err = downloader.DownloadFile(db, fullURL)
		if err != nil {
			return err
		}
		if s.Verbose {
			fmt.Println(time.Now().Format("2006-01-02 03:04:05"))
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
	filePath := filepath.Join(s.Config.Main.CacheDir, "files/company_tickers.json")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

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
		fmt.Print("Indexing file company_tickers.json: ")
	}

	for _, v := range allCompanyTickers {
		err = SaveCIK(db, v.Cik)
		if err != nil {
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
			return err
		}
	}
	if s.Verbose {
		fmt.Println("\u2713")
	}
	return nil
}

func (s *SEC) ExchangeTickersGet(db *sqlx.DB) error {
	// Retrieving JSON data from URL
	filePath := filepath.Join(s.Config.Main.CacheDir, "files/company_tickers_exchange.json")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

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
		fmt.Print("Indexing file company_tickers_exchange.json: ")
	}

	for _, v := range fileExchange.Data {
		err = SaveCIK(db, int(v[0].(float64)))
		if err != nil {
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
			return err
		}
	}
	if s.Verbose {
		fmt.Println("\u2713")
	}
	return nil
}

func (s *SEC) ParseRSSGoXML(path string) (RSSFile, error) {
	var rssFile RSSFile

	xmlFile, err := os.Open(path)
	if err != nil {
		return rssFile, err
	}

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
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

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
			fmt.Printf("Checking file '%v' in disk: ", filepath.Base(fileURL))
		}
		isFileCorrect, err := downloader.FileCorrect(db, fileURL)
		if err != nil {
			return err
		}
		if s.Verbose && isFileCorrect {
			fmt.Println("\u2713")
		}

		if !isFileCorrect {
			if s.Verbose {
				fmt.Print("Downloading file...: ")
			}
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			if s.Verbose {
				fmt.Println(time.Now().Format("2006-01-02 03:04:05"))
			}
			time.Sleep(rateLimit)
		}
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
	DO UPDATE SET will_download=EXCLUDED.will_download, updated_at=NOW() 
	WHERE worklist.year=EXCLUDED.year AND worklist.month=EXCLUDED.month ;`, year, month, willDownload)
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

		filePath := filepath.Join(s.Config.Main.CacheDir, fileUrl.Path)
		_, err = os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("inserted into database all downloaded files, run sec dow data then run sec index again to insert all enabled months/years")
		}

		xbrlFile, err := os.Open(filePath)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadAll(xbrlFile)
		if err != nil {
			return err
		}

		var fileBody string
		if s.IsFileIndexable(filePath) {
			fileBody, err = html2text.FromString(string(data))
			if err != nil {
				return err
			}
		}

		_, err = db.Exec(`
		INSERT INTO sec.secItemFile (title, link, guid, enclosure_url, enclosure_length, enclosure_type, description, pubdate, companyname, formtype, fillingdate, ciknumber, accessionnumber, filenumber, acceptancedatetime, period, assistantdirector, assignedsic, fiscalyearend, xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl, xbrlbody, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, NOW(), NOW()) 

		ON CONFLICT (xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl)
		DO UPDATE SET title=EXCLUDED.title, link=EXCLUDED.link, guid=EXCLUDED.guid, enclosure_url=EXCLUDED.enclosure_url, enclosure_length=EXCLUDED.enclosure_length, enclosure_type=EXCLUDED.enclosure_type, description=EXCLUDED.description, pubdate=EXCLUDED.pubdate, companyname=EXCLUDED.companyname, formtype=EXCLUDED.formtype, fillingdate=EXCLUDED.fillingdate, ciknumber=EXCLUDED.ciknumber, accessionnumber=EXCLUDED.accessionnumber, filenumber=EXCLUDED.filenumber, acceptancedatetime=EXCLUDED.acceptancedatetime, period=EXCLUDED.period, assistantdirector=EXCLUDED.assistantdirector, assignedsic=EXCLUDED.assignedsic, fiscalyearend=EXCLUDED.fiscalyearend, xbrlsequence=EXCLUDED.xbrlsequence, xbrlfile=EXCLUDED.xbrlfile, xbrltype=EXCLUDED.xbrltype, xbrlsize=EXCLUDED.xbrlsize, xbrldescription=EXCLUDED.xbrldescription, xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl, xbrlurl=EXCLUDED.xbrlurl, updated_at=NOW()
		WHERE secItemFile.xbrlsequence=EXCLUDED.xbrlsequence AND secItemFile.xbrlfile=EXCLUDED.xbrlfile AND secItemFile.xbrltype=EXCLUDED.xbrltype AND secItemFile.xbrlsize=EXCLUDED.xbrlsize AND secItemFile.xbrldescription=EXCLUDED.xbrldescription AND secItemFile.xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl AND secItemFile.xbrlurl=EXCLUDED.xbrlurl AND secItemFile.xbrlbody=EXCLUDED.xbrlbody;`,
			item.Title, item.Link, item.Guid, item.Enclosure.URL, enclosureLength, item.Enclosure.Type, item.Description, item.PubDate, item.XbrlFiling.CompanyName, item.XbrlFiling.FormType, item.XbrlFiling.FilingDate, cikNumber, item.XbrlFiling.AccessionNumber, item.XbrlFiling.FileNumber, item.XbrlFiling.AcceptanceDatetime, item.XbrlFiling.Period, item.XbrlFiling.AssistantDirector, assignedSic, fiscalYearEnd, xbrlSequence, v.File, v.Type, xbrlSize, v.Description, xbrlInline, v.URL, fileBody)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SEC) IsFileIndexable(filename string) bool {
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
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	for _, v := range files {
		isFileCorrect, err := downloader.FileCorrect(db, v.URL)
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
		if !s.Verbose {
			fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", *currentCount, totalCount, (totalCount - *currentCount))
		}

		if s.Verbose {
			fmt.Printf("[%d/%d] %s downloaded...\n", *currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
		}
		time.Sleep(rateLimit)
	}

	return nil
}

func CheckRSSAvailability(year int, month int) (err error) {
	if year < XMLStartYear {
		err = fmt.Errorf("the earliest available XML is %d/%d", XMLStartYear, XMLStartMonth)
		return err
	}

	if year == XMLStartYear && month > 0 && month < XMLStartMonth {
		err = fmt.Errorf("the earliest available XML is %d/%d", XMLStartYear, XMLStartMonth)
		return err
	}

	if year > time.Now().Year() || month < 0 || month > 12 || (year == time.Now().Year() && month > int(time.Now().Month())) {
		err = fmt.Errorf("the latest available XML is %d/%d", time.Now().Year(), time.Now().Month())
		return err
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

	for _, file := range files {
		reader, err := file.Open()
		if err != nil {
			return err
		}

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
			return err
		}
	}
	return nil
}

func (s *SEC) SearchByFillingDate(db *sqlx.DB, startdate string, enddate string) ([]SECItemFile, error) {
	secItemFiles := []SECItemFile{}
	err := db.Select(&secItemFiles, `
	SELECT sec.tickers.ticker, sec.secItemFile.title, sec.secItemFile.companyname, sec.secItemFile.ciknumber, sec.secItemFile. accessionnumber, sec.secItemFile.xbrlfile 
	FROM sec.secItemFile 
	LEFT JOIN sec.tickers
	ON sec.secitemfile.ciknumber = sec.tickers.cik
	WHERE DATE(fillingdate) between $1 AND $2
	`, startdate, enddate)
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
			out, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer out.Close()

			reader, err := file.Open()
			if err != nil {
				return err
			}

			buf := bytes.Buffer{}
			_, err = buf.ReadFrom(reader)
			if err != nil {
				return err
			}

			_, err = io.Copy(out, &buf)
			if err != nil {
				return err
			}
		}
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
		fmt.Print("Calculating number of XBRL Files in the index files: ")
	}

	totalCount, err := s.TotalXbrlFileCountGet(worklist, s.Config.Main.CacheDir)
	if err != nil {
		return err
	}
	if s.Verbose {
		fmt.Println(totalCount)
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

		rssFile, err := s.ParseRSSGoXML(fileURL)
		if err != nil {
			return err
		}

		if s.Verbose {
			fmt.Println(verboseMessage)
		}

		err = implementFunc(db, rssFile, worklist)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *SEC) DownloadZIPFiles(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
	downloader := download.NewDownloader(s.Config)
	downloader.Verbose = s.Verbose
	downloader.Debug = s.Debug

	rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", s.Config.Main.RateLimitMs))
	if err != nil {
		return err
	}

	totalCount := len(rssFile.Channel.Item)
	currentCount := 0
	for _, v1 := range rssFile.Channel.Item {
		if v1.Enclosure.URL != "" {

			isFileCorrect, err := downloader.FileCorrect(db, v1.Enclosure.URL)
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
			if !s.Verbose {
				fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", currentCount, totalCount, (totalCount - currentCount))
			}

			if s.Verbose {
				fmt.Printf("[%d/%d] %s downloaded...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
			}
		}
	}
	return nil
}

func (s *SEC) IndexZIPFileContent(db *sqlx.DB, rssFile RSSFile, worklist []Worklist) error {
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
			return fmt.Errorf("please run sec dowz to download all ZIP files then run sec indexz again to index them")
		}

		reader, err := zip.OpenReader(zipCachePath)
		if err != nil {
			return err
		}

		defer reader.Close()

		err = s.ZIPContentUpsert(db, zipPath, reader.File)
		if err != nil {
			return err
		}
		currentCount++

		if s.Verbose {
			fmt.Printf("[%d/%d] %s inserted for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
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
			return fmt.Errorf("please run sec dowz to download all ZIP files then run sec indexz again to index them")
		}

		reader, err := zip.OpenReader(zipCachePath)
		if err != nil {
			return err
		}

		defer reader.Close()

		err = s.CreateFilesFromZIP(zipPath, reader.File)
		if err != nil {
			return err
		}

		currentCount++
		if s.Verbose {
			fmt.Printf("[%d/%d] %s unpacked...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
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
			fmt.Printf("[%d/%d] %s inserted for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
		}
	}
	return nil
}
