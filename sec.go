// Copyright (c) 2021 Equres LLC. All rights reserved.

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
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

// Parsing RSS/XML using GoFeed
func (s *SEC) ParseRSSGoFeed(url string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(s.BaseURL + url)
	if err != nil {
		return err
	}

	for _, v := range feed.Items {
		for _, v1 := range v.Extensions {
			for _, v2 := range v1["xbrlFiling"] {
				for _, v3 := range v2.Children["xbrlFiles"][0].Children["xbrlFile"] {
					err = s.DownloadFile(url, v3.Attrs["url"])
					if err != nil {
						return nil
					}
					time.Sleep(1 * time.Second)
				}
			}
		}
	}
	return nil
}

// Parsing RSS/XML using Go XML Library
func (s *SEC) ParseRSSGoXML(url string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", s.BaseURL+url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Equres LLC wojciech@koszek.com")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rssFile RSSFile

	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&rssFile)
	if err != nil {
		return err
	}

	for _, v := range rssFile.Channel.Item[:1] {
		for _, v1 := range v.XbrlFiling.XbrlFiles.XbrlFile {
			err = s.DownloadFile(url, v1.URL)
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
		}
	}

	return err
}

func (s *SEC) DownloadFile(basepath, fullurl string) error {
	resp, err := http.Get(fullurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = os.Stat(basepath)
	if err != nil {
		err = os.MkdirAll(basepath, 0755)
		if err != nil {
			return err
		}
	}

	out, err := os.Create(strings.Join([]string{basepath, filepath.Base(fullurl)}, "/"))
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
