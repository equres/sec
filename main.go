package main

import (
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	db, err := ConnectDB()
	if err != nil {
		panic(err)
	}

	body, err := FetchFile("https://www.sec.gov/files/company_tickers.json")
	if err != nil {
		panic(err)
	}

	// Creating Map to hold company  ticker structs
	allCompanyTickers := make(map[int]SecTicker)

	// Converting JSON to Structs
	json.Unmarshal(body, &allCompanyTickers)

	// Insert into DB
	for _, v := range allCompanyTickers {
		sec := NewSEC(v.Cik, v.Ticker, v.Title, v.Exchange)
		err := sec.SaveSecTicker(db)
		if err != nil {
			panic(err)
		}
	}

	// Retrieving all SecTickers
	tickers, err := GetAllTickers(db)
	if err != nil {
		panic(err)
	}

	fmt.Println(tickers)
}
