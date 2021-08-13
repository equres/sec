// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

func (s *SEC) DownloadFile(fullurl string, cfg Config) error {
	filePath := strings.ReplaceAll(fullurl, s.BaseURL, "")
	cachePath := fmt.Sprintf("%v%v", cfg.CacheDir, filePath)

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

	size, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	filestat, err := os.Stat(cachePath)
	if err != nil || filestat.Size() != size {
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

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		return nil
	}
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
