// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	sec := NewSEC("https://www.sec.gov/")

	db, err := ConnectDB()
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 1 {
		for _, v := range os.Args[1:] {
			switch v {
			case "update":
				// Update/Insert data into DB
				err = sec.TickerUpdateAll(db)
				if err != nil {
					panic(err)
				}
			case "retrieve":
				// Retrieving all SecTickers
				tickers, err := sec.TickersGetAll(db)
				if err != nil {
					panic(err)
				}
				fmt.Println(tickers)
			default:
				fmt.Println("Add an expression to run using CLI. \n'update' to update the list from the website\n'retrieve' to display the list in the command-line")
			}
		}
	}
}
