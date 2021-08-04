// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	db, err := ConnectDB()
	if err != nil {
		panic(err)
	}

	sec := NewSEC("https://www.sec.gov/")

	// Retrieving JSON data from URL
	body, err := sec.FetchFile("files/company_tickers.json")
	if err != nil {
		panic(err)
	}

	// Update/Insert data into DB
	err = sec.TickerUpdateAll(db, body)
	if err != nil {
		panic(err)
	}

	// Retrieving all SecTickers
	tickers, err := sec.TickersGetAll(db)
	if err != nil {
		panic(err)
	}

	fmt.Println(tickers)
}
