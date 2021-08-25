// Copyright (c) 2021 Equres LLC. All rights reserved.

package sec

import (
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
	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
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
	Year          int
	Month         int
	Will_download bool
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

func (s *SEC) FetchFile(url_var string) ([]byte, error) {
	// Retrieving JSON From URL
	// main_url := path.Join(s.BaseURL, url_var)
	parsed_url, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, err
	}
	other_url, err := url.Parse(url_var)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(parsed_url.ResolveReference(other_url).String())
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
	// Retrieving JSON data from URL
	body, err := s.FetchFile("files/company_tickers.json")
	if err != nil {
		return err
	}

	// Creating Map to hold company ticker structs
	allCompanyTickers := make(map[int]SecTicker)

	// Converting JSON to Structs
	err = json.Unmarshal(body, &allCompanyTickers)
	if err != nil {
		return err
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
	return nil
}

func (s *SEC) ExchangeTickersGet(db *sqlx.DB) error {
	// Retrieving JSON data from URL
	body, err := s.FetchFile("files/company_tickers_exchange.json")
	if err != nil {
		return err
	}

	fileExchange := ExchangesFile{}

	err = json.Unmarshal(body, &fileExchange)
	if err != nil {
		return err
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

func (s *SEC) DownloadIndex() error {
	db, err := database.ConnectDB(s.Config)
	if err != nil {
		return err
	}

	worklist, err := WorklistWillDownloadGet(db)
	if err != nil {
		return err
	}

	downloader := download.NewDownloader(s.Config)
	downloader.RateLimitDuration = 1 * time.Second

	for _, v := range worklist {
		date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
		if err != nil {
			return err
		}
		formatted := date.Format("2006-01")

		fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", s.BaseURL, formatted)

		if s.Verbose {
			fmt.Printf("Checking file 'xbrlrss-%v.xml' in disk: ", formatted)
		}
		not_download, err := downloader.FileInCache(fileURL)
		if err != nil {
			return err
		}
		if s.Verbose && not_download {
			fmt.Println("\u2713")
		}

		if !not_download {
			if s.Verbose {
				fmt.Printf("File 'xbrlrss-%v.xml' is not in disk. Downloading file...: ", formatted)
			}
			err = downloader.DownloadFile(db, fileURL)
			if err != nil {
				return err
			}
			if s.Verbose {
				fmt.Println(time.Now().Format("2006-01-02 03:04:05"))
			}
			time.Sleep(downloader.RateLimitDuration)
		}
	}
	return nil
}

func (s *SEC) CalculateRSSFilesZIP(rssFile RSSFile) (int, error) {
	var total_size int
	for _, v := range rssFile.Channel.Item {
		if v.Enclosure.Length != "" {
			val, err := strconv.Atoi(v.Enclosure.Length)
			if err != nil {
				return 0, err
			}
			total_size += val
		}
	}
	return total_size, nil
}

func SaveWorklist(year int, month int, will_download bool, db *sqlx.DB) error {
	_, err := db.Exec(`
	INSERT INTO sec.worklist (year, month, will_download, created_at, updated_at) 
	VALUES ($1, $2, $3, NOW(), NOW()) 
	ON CONFLICT (month, year) 
	DO UPDATE SET will_download=EXCLUDED.will_download, updated_at=NOW() 
	WHERE worklist.year=EXCLUDED.year AND worklist.month=EXCLUDED.month ;`, year, month, will_download)
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

		filePath := strings.ReplaceAll(v.URL, s.BaseURL, s.Config.Main.CacheDir)

		_, err = os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("successfully inserted into database all downloaded files")
		}

		xbrlFile, err := os.Open(filePath)
		if err != nil {
			return err
		}

		fileBody, err := ioutil.ReadAll(xbrlFile)
		if err != nil {
			return err
		}

		// Skip saving image body in DB
		fileExtension := filepath.Ext(filePath)
		if fileExtension == ".jpg" || fileExtension == ".jpeg" {
			fileBody = []byte{}
		}

		_, err = db.Exec(`
		INSERT INTO sec.secItemFile (title, link, guid, enclosure_url, enclosure_length, enclosure_type, description, pubdate, companyname, formtype, fillingdate, ciknumber, accessionnumber, filenumber, acceptancedatetime, period, assistantdirector, assignedsic, fiscalyearend, xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl, xbrlbody, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, NOW(), NOW()) 

		ON CONFLICT (xbrlsequence, xbrlfile, xbrltype, xbrlsize, xbrldescription, xbrlinlinexbrl, xbrlurl)
		DO UPDATE SET title=EXCLUDED.title, link=EXCLUDED.link, guid=EXCLUDED.guid, enclosure_url=EXCLUDED.enclosure_url, enclosure_length=EXCLUDED.enclosure_length, enclosure_type=EXCLUDED.enclosure_type, description=EXCLUDED.description, pubdate=EXCLUDED.pubdate, companyname=EXCLUDED.companyname, formtype=EXCLUDED.formtype, fillingdate=EXCLUDED.fillingdate, ciknumber=EXCLUDED.ciknumber, accessionnumber=EXCLUDED.accessionnumber, filenumber=EXCLUDED.filenumber, acceptancedatetime=EXCLUDED.acceptancedatetime, period=EXCLUDED.period, assistantdirector=EXCLUDED.assistantdirector, assignedsic=EXCLUDED.assignedsic, fiscalyearend=EXCLUDED.fiscalyearend, xbrlsequence=EXCLUDED.xbrlsequence, xbrlfile=EXCLUDED.xbrlfile, xbrltype=EXCLUDED.xbrltype, xbrlsize=EXCLUDED.xbrlsize, xbrldescription=EXCLUDED.xbrldescription, xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl, xbrlurl=EXCLUDED.xbrlurl, updated_at=NOW()
		WHERE secItemFile.xbrlsequence=EXCLUDED.xbrlsequence AND secItemFile.xbrlfile=EXCLUDED.xbrlfile AND secItemFile.xbrltype=EXCLUDED.xbrltype AND secItemFile.xbrlsize=EXCLUDED.xbrlsize AND secItemFile.xbrldescription=EXCLUDED.xbrldescription AND secItemFile.xbrlinlinexbrl=EXCLUDED.xbrlinlinexbrl AND secItemFile.xbrlurl=EXCLUDED.xbrlurl AND secItemFile.xbrlbody=EXCLUDED.xbrlbody;`,
			item.Title, item.Link, item.Guid, item.Enclosure.URL, enclosureLength, item.Enclosure.Type, item.Description, item.PubDate, item.XbrlFiling.CompanyName, item.XbrlFiling.FormType, item.XbrlFiling.FilingDate, item.XbrlFiling.CikNumber, item.XbrlFiling.AccessionNumber, item.XbrlFiling.FileNumber, item.XbrlFiling.AcceptanceDatetime, item.XbrlFiling.Period, item.XbrlFiling.AssistantDirector, assignedSic, fiscalYearEnd, xbrlSequence, v.File, v.Type, xbrlSize, v.Description, xbrlInline, v.URL, string(fileBody))
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseYearMonth(year_month string) (year int, month int, err error) {
	switch len(year_month) {
	case 4:
		date, err := time.Parse("2006", year_month)
		if err != nil {
			return 0, 0, err
		}
		year = date.Year()
	case 6:
		date, err := time.Parse("2006/1", year_month)
		if err != nil {
			return 0, 0, err
		}
		year = date.Year()
		month = int(date.Month())
	case 7:
		date, err := time.Parse("2006/01", year_month)
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

func (s *SEC) TotalXbrlFileCountGet(worklist []Worklist, cache_dir string) (int, error) {
	var total_count int
	for _, v := range worklist {
		date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
		if err != nil {
			return 0, err
		}
		formatted := date.Format("2006-01")

		filepath := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", cache_dir, formatted)
		rssFile, err := s.ParseRSSGoXML(filepath)
		if err != nil {
			return 0, err
		}

		for _, v1 := range rssFile.Channel.Item {
			total_count += len(v1.XbrlFiling.XbrlFiles.XbrlFile)
		}
	}
	return total_count, nil
}

func (s *SEC) DownloadXbrlFileContent(files []XbrlFile, config config.Config, current_count *int, total_count int) error {
	db, err := database.ConnectDB(s.Config)
	if err != nil {
		return err
	}

	downloader := download.NewDownloader(s.Config)
	downloader.RateLimitDuration = 1 * time.Second

	for _, v := range files {
		not_download, err := downloader.FileInCache(v.URL)
		if err != nil {
			return err
		}

		if !not_download {
			err = downloader.DownloadFile(db, v.URL)
			if err != nil {
				return err
			}
			time.Sleep(downloader.RateLimitDuration)
		}

		*current_count++
		if !s.Verbose {
			fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", *current_count, total_count, (total_count - *current_count))
		}

		if s.Verbose {
			fmt.Printf("[%d/%d] %s downloaded...\n", *current_count, total_count, time.Now().Format("2006-01-02 03:04:05"))
		}
		time.Sleep(1 * time.Second)
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

func (s *SEC) Downloadability(year int, month int, will_download bool) error {
	db, err := database.ConnectDB(s.Config)
	if err != nil {
		return err
	}

	if month != 0 {
		err = SaveWorklist(year, month, will_download, db)
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
		err = SaveWorklist(year, i, will_download, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func WorklistWillDownloadGet(db *sqlx.DB) ([]Worklist, error) {
	// Retrieve from DB
	var worklist []Worklist

	err := db.Select(&worklist, "SELECT year, month, will_download FROM sec.worklist WHERE will_download = true")
	if err != nil {
		return nil, err
	}
	return worklist, nil
}
