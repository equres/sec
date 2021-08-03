package main

import (
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

func (t SecTicker) String() string {
	return fmt.Sprintf("Cik: %d\nTicker: %s\nTitle: %s\nExchange: %s\n", t.Cik, t.Ticker, t.Title, t.Exchange)
}

func NewSEC(cik int, ticker, title, exchange string) *SecTicker {
	return &SecTicker{
		Cik:      cik,
		Ticker:   ticker,
		Title:    title,
		Exchange: exchange,
	}
}

func (t SecTicker) SaveSecTicker(db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO tickers (cik, ticker, title, exchange, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW());`, t.Cik, t.Ticker, t.Title, t.Exchange)
	if err != nil {
		return err
	}
	return nil
}

func FetchFile(url string) ([]byte, error) {
	// Retrieving JSON From URL
	resp, err := http.Get(url)
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

func GetAllTickers(db *sqlx.DB) ([]SecTicker, error) {
	// Retrieve from DB
	tickers := []SecTicker{}
	err := db.Select(&tickers, "SELECT cik, ticker, title, exchange FROM tickers")
	if err != nil {
		return nil, err
	}
	return tickers, nil
}
