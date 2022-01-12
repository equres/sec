package secticker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/seccik"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// Ticker Struct Based on JSON
type SecTicker struct {
	Cik      int    `json:"cik_str"`
	Ticker   string `json:"ticker"`
	Title    string `json:"title"`
	Exchange string `json:"exchange"`
}

type ExchangesFile struct {
	Fields []string        `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

func (t SecTicker) String() string {
	return fmt.Sprintf("Cik: %d\nTicker: %s\nTitle: %s\nExchange: %s\n", t.Cik, t.Ticker, t.Title, t.Exchange)
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

func UpdateAll(s *sec.SEC, db *sqlx.DB) error {
	err := NoExchangeFileGet(s, db)
	if err != nil {
		return err
	}

	err = ExchangeFileGet(s, db)
	if err != nil {
		return err
	}
	return nil
}

func NoExchangeFileGet(s *sec.SEC, db *sqlx.DB) error {
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
		err = seccik.SaveCIK(db, v.Cik)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers.json", "failed", "error_inserting_cik_in_database")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	for _, v := range allCompanyTickers {
		ticker := SecTicker{
			Cik:      v.Cik,
			Ticker:   v.Ticker,
			Title:    v.Title,
			Exchange: v.Exchange,
		}
		err := ticker.Save(db)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers.json", "failed", "error_inserting_ticker_in_database")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	if s.Verbose {
		log.Info("\u2713")
	}
	eventErr := database.CreateIndexEvent(db, "company_tickers.json", "success", "")
	if eventErr != nil {
		return eventErr
	}

	return nil
}

func ExchangeFileGet(s *sec.SEC, db *sqlx.DB) error {
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
		err = seccik.SaveCIK(db, int(v[0].(float64)))
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "failed", "error_inserting_cik_in_database")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}

	for _, v := range fileExchange.Data {
		// Below is because sometimes the exchange is empty (nil). Added lines to ensure no error when saving
		cik := 0
		if v[0] != nil {
			cik = int(v[0].(float64))
		}

		title := ""
		if v[1] != nil {
			title = v[1].(string)
		}

		ticker := ""
		if v[2] != nil {
			ticker = v[2].(string)
		}

		exchange := ""
		if v[3] != nil {
			exchange = v[3].(string)
		}

		sec := SecTicker{
			Cik:      cik,
			Title:    title,
			Ticker:   ticker,
			Exchange: exchange,
		}
		err := sec.Save(db)
		if err != nil {
			eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "failed", "error_inserting_ticker_in_database")
			if eventErr != nil {
				return eventErr
			}
			return err
		}
	}
	if s.Verbose {
		log.Info("\u2713")
	}
	eventErr := database.CreateIndexEvent(db, "company_tickers_exchange.json", "success", "")
	if eventErr != nil {
		return eventErr
	}
	return nil
}

func GetAll(db *sqlx.DB) ([]SecTicker, error) {
	// Retrieve from DB
	tickers := []SecTicker{}
	err := db.Select(&tickers, "SELECT cik, ticker, title, exchange FROM sec.tickers")
	if err != nil {
		return nil, err
	}
	return tickers, nil
}
