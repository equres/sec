package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
)

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
		Tickers: []SecTicker{},
	}
}

func (t SecTicker) Save(db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO tickers (cik, ticker, title, exchange, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) ON CONFLICT (cik, ticker, title, exchange) DO UPDATE SET cik=EXCLUDED.cik, ticker=EXCLUDED.ticker, title=EXCLUDED.title, exchange=EXCLUDED.exchange, updated_at=NOW() WHERE tickers.cik = EXCLUDED.cik;`, t.Cik, t.Ticker, t.Title, t.Exchange)
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

func (s *SEC) TickerUpdateAll(db *sqlx.DB, body []byte) error {
	// Creating Map to hold company  ticker structs
	allCompanyTickers := make(map[int]SecTicker)

	// Converting JSON to Structs
	json.Unmarshal(body, &allCompanyTickers)

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

func (s *SEC) GetAllTickers(db *sqlx.DB) ([]SecTicker, error) {
	// Retrieve from DB
	tickers := []SecTicker{}
	err := db.Select(&tickers, "SELECT cik, ticker, title, exchange FROM tickers")
	if err != nil {
		return nil, err
	}
	return tickers, nil
}
