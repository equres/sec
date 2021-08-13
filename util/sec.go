// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/html/charset"
)

type RSSFile struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
			Atom string `xml:"atom,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Language      string `xml:"language"`
		PubDate       string `xml:"pubDate"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text      string `xml:",chardata"`
			Title     string `xml:"title"`
			Link      string `xml:"link"`
			Guid      string `xml:"guid"`
			Enclosure struct {
				Text   string `xml:",chardata"`
				URL    string `xml:"url,attr"`
				Length string `xml:"length,attr"`
				Type   string `xml:"type,attr"`
			} `xml:"enclosure"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			XbrlFiling  struct {
				Text               string `xml:",chardata"`
				Edgar              string `xml:"edgar,attr"`
				CompanyName        string `xml:"companyName"`
				FormType           string `xml:"formType"`
				FilingDate         string `xml:"filingDate"`
				CikNumber          string `xml:"cikNumber"`
				AccessionNumber    string `xml:"accessionNumber"`
				FileNumber         string `xml:"fileNumber"`
				AcceptanceDatetime string `xml:"acceptanceDatetime"`
				Period             string `xml:"period"`
				AssistantDirector  string `xml:"assistantDirector"`
				AssignedSic        string `xml:"assignedSic"`
				FiscalYearEnd      string `xml:"fiscalYearEnd"`
				XbrlFiles          struct {
					Text     string `xml:",chardata"`
					XbrlFile []struct {
						Text        string `xml:",chardata"`
						Sequence    string `xml:"sequence,attr"`
						File        string `xml:"file,attr"`
						Type        string `xml:"type,attr"`
						Size        string `xml:"size,attr"`
						Description string `xml:"description,attr"`
						InlineXBRL  string `xml:"inlineXBRL,attr"`
						URL         string `xml:"url,attr"`
					} `xml:"xbrlFile"`
				} `xml:"xbrlFiles"`
			} `xml:"xbrlFiling"`
		} `xml:"item"`
	} `xml:"channel"`
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
}

func (t SecTicker) String() string {
	return fmt.Sprintf("Cik: %d\nTicker: %s\nTitle: %s\nExchange: %s\n", t.Cik, t.Ticker, t.Title, t.Exchange)
}

func NewSEC(baseUrl string) *SEC {
	return &SEC{
		BaseURL: baseUrl,
	}
}

func (t SecTicker) Save(db *sqlx.DB) error {
	_, err := db.Exec(`
	INSERT INTO tickers (cik, ticker, title, exchange, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, NOW(), NOW()) 
	ON CONFLICT (cik, ticker, title) 
	DO UPDATE SET cik=EXCLUDED.cik, ticker=EXCLUDED.ticker, title=EXCLUDED.title, exchange=EXCLUDED.exchange, updated_at=NOW() 
	WHERE tickers.cik=EXCLUDED.cik AND tickers.ticker=EXCLUDED.ticker AND tickers.title=EXCLUDED.title AND tickers.exchange = '';`, t.Cik, t.Ticker, t.Title, t.Exchange)
	if err != nil {
		return err
	}
	return nil
}

func (s *SEC) FetchFile(url string) ([]byte, error) {
	// Retrieving JSON From URL
	resp, err := http.Get(s.BaseURL + url)

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
	err := db.Select(&tickers, "SELECT cik, ticker, title, exchange FROM tickers")
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

func (s *SEC) ParseRSSGoXML(url string) (RSSFile, error) {
	client := &http.Client{}
	var rssFile RSSFile

	req, err := http.NewRequest("GET", s.BaseURL+url, nil)
	if err != nil {
		return rssFile, err
	}
	req.Header.Set("User-Agent", "Equres LLC wojciech@koszek.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := client.Do(req)
	if err != nil {
		return rssFile, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rssFile, err
	}

	reader := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return rssFile, err
	}

	decoder := xml.NewDecoder(gzipReader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&rssFile)
	if err != nil {
		return rssFile, err
	}
	return rssFile, nil
}

// Parsing RSS/XML using Go XML Library
func (s *SEC) DownloadXbrlFiles(rssFile RSSFile, basepath string, isVerbose bool) error {
	var total_count int
	var current_count int
	if !isVerbose {
		for _, v := range rssFile.Channel.Item {
			total_count += len(v.XbrlFiling.XbrlFiles.XbrlFile)
		}

		fmt.Printf("[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report\n", current_count, total_count, (total_count - current_count))
	}

	for _, v := range rssFile.Channel.Item {
		for _, v1 := range v.XbrlFiling.XbrlFiles.XbrlFile {
			size, err := strconv.ParseFloat(v1.Size, 64)
			if err != nil {
				return err
			}

			err = s.DownloadFile(basepath, v1.URL, size)
			if err != nil {
				return err
			}

			current_count++
			if !isVerbose {
				fmt.Printf("[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report\n", current_count, total_count, (total_count - current_count))
			}

			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (s *SEC) CreateFile(basepath, fullurl string, body io.Reader) error {
	out, err := os.Create(strings.Join([]string{basepath, filepath.Base(fullurl)}, "/"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, body)
	if err != nil {
		return err
	}

	return nil
}

func (s *SEC) DownloadFile(basepath, fullurl string, size float64) error {
	// Base path is path from current folder to download folder
	// Fullurl is the actual URL for the file to be downloaded
	client := &http.Client{}

	req, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Equres LLC wojciech@koszek.com")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Check if Folder for this XML exists
	_, err = os.Stat(basepath)
	if err != nil {
		err = os.MkdirAll(basepath, 0755)
		if err != nil {
			return err
		}
	}

	filestat, err := os.Stat(basepath + "/" + filepath.Base(fullurl))

	if err != nil {
		err = s.CreateFile(basepath, fullurl, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}

	if filestat.Size() != int64(size) {
		// fmt.Println("Old File Exists: Updating: " + basepath + "/" + filepath.Base(fullurl))
		err = s.CreateFile(basepath, fullurl, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}

	// fmt.Println("File already exists: " + basepath + "/" + filepath.Base(fullurl))
	return nil
}

func SaveWorklist(year int, month int, will_download bool, db *sqlx.DB) error {
	_, err := db.Exec(`
	INSERT INTO worklist (year, month, will_download, created_at, updated_at) 
	VALUES ($1, $2, $3, NOW(), NOW()) 
	ON CONFLICT (month, year) 
	DO UPDATE SET will_download=EXCLUDED.will_download, updated_at=NOW() 
	WHERE worklist.year=EXCLUDED.year AND worklist.month=EXCLUDED.month ;`, year, month, will_download)
	if err != nil {
		return err
	}
	return nil
}

func Downloadability(year_month string, will_download bool) error {
	var month int
	var year int

	switch len(year_month) {
	case 4:
		date, err := time.Parse("2006", year_month)
		if err != nil {
			return err
		}
		year = date.Year()
	case 6:
		date, err := time.Parse("2006/1", year_month)
		if err != nil {
			return err
		}
		year = date.Year()
		month = int(date.Month())
	case 7:
		date, err := time.Parse("2006/01", year_month)
		if err != nil {
			return err
		}
		year = date.Year()
		month = int(date.Month())
	default:
		err := errors.New("please enter a valid date ('2021' or '2021/05')")
		return err
	}

	db, err := ConnectDB()
	if err != nil {
		panic(err)
	}

	if month != 0 {
		err = SaveWorklist(year, month, will_download, db)
		if err != nil {
			panic(err)
		}
		return nil
	}

	for i := 1; i <= 12; i++ {
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

	err := db.Select(&worklist, "SELECT year, month, will_download FROM worklist WHERE will_download = true")
	if err != nil {
		return nil, err
	}
	return worklist, nil
}
