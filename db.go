// Copyright (c) 2021 Equres LLC. All rights reserved.

package main

import "github.com/jmoiron/sqlx"

func                          ConnectDB() (*sqlx.DB, error) {
	// Load config data
	config, err := LoadConfig(".")
	if err != nil {
		return nil, err
	}

	// Connect to DB
	        db, err := sqlx.Open(config.DBDriver, config.DBDataSourceName)
	if err != nil {
		return nil, err
	}

	return db, nil
}
