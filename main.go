package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

// DB Information
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "hazem1999"
	dbname   = "financial_research_website"
)

// Ticker Struct Based on JSON
type JsonCompanyTicker struct {
	Cik    int    `json:"cik_str"`
	Ticker string `json:"ticker"`
	Title  string `json:"title"`
}

// Ticker Struct After Receiving From DB
type Ticker struct {
	ID         int            `json:"id"`
	Cik        int            `json:"cik"`
	Ticker     string         `json:"ticker"`
	Name       string         `json:"name"`
	Created_at sql.NullString `json:"created_at"`
	Updated_at sql.NullString `json:"updated_at"`
	Deleted_at sql.NullString `json:"deleted_at"`
}

func (t Ticker) String() string {
	return fmt.Sprintf("ID: %d\nCik: %d\nTicker: %s\nName: %s\n", t.ID, t.Cik, t.Ticker, t.Name)
}

// Function to connect to Database
func ConnectDB(host, user, password, dbname string, port int) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	// defer db.Close()

	return db, nil
}

func main() {
	// Connect to database
	db, err := ConnectDB(host, user, password, dbname, port)
	if err != nil {
		panic(err)
	}

	// Retrieving JSON From URL
	resp, err := http.Get("https://www.sec.gov/files/company_tickers.json")
	if err != nil {
		panic(err)
	}

	// Reading JSON data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Creating Map to hold company  ticker structs
	allCompanyTickers := make(map[int]JsonCompanyTicker)

	// Converting JSON to Structs
	json.Unmarshal(body, &allCompanyTickers)

	// COMMENTED OUT BECAUSE IT IS ALREADY INSERTED
	// Insert Data Into DB
	// err = MultiInsertDB(db, allCompanyTickers)
	// if err != nil {
	// 	panic(err)
	// }

	// Retrieve from DB
	rows, err := db.Query("SELECT * FROM tickers")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Adding rows of DB to slice
	tickers := make([]Ticker, 0)

	for rows.Next() {
		tick := Ticker{}
		err := rows.Scan(&tick.ID, &tick.Ticker, &tick.Cik, &tick.Name, &tick.Created_at, &tick.Updated_at, &tick.Deleted_at)
		if err != nil {
			panic(err)
		}
		tickers = append(tickers, tick)
	}

	fmt.Println(tickers[0].String())

}

// Insert JsonTicker Map to DB
func MultiInsertDB(db *sql.DB, tickers map[int]JsonCompanyTicker) error {
	sqlStatement := `
	INSERT INTO tickers (ticker, cik, name, created_at, updated_at)
	VALUES ($1, $2, $3, NOW(), NOW());`
	for _, v := range tickers {
		_, err := db.Exec(sqlStatement, v.Ticker, v.Cik, v.Title)
		if err != nil {
			return err
		}
	}
	return nil
}
