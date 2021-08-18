// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/johejo/golang-migrate-extra/source/iofs"
)

func ConnectDB() (*sqlx.DB, error) {
	// Load config data
	config, err := LoadConfig("./ci")
	if err != nil {
		return nil, err
	}

	// Connect to DB
	db, err := sqlx.Open(config.Database.Driver, config.DBGetDataSourceName())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateUp(db *sqlx.DB, fs embed.FS) error {
	config, err := LoadConfig("./ci")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		fmt.Println("failed to make migrations")
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DBGetURL())
	if err != nil {
		fmt.Println("failed make iofs source")
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		fmt.Println("failed to UP the migrations")
		return err
	}

	return nil
}

func MigrateDown(db *sqlx.DB, fs embed.FS) error {
	config, err := LoadConfig("./ci")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		fmt.Println("failed to make migrations")
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DBGetURL())
	if err != nil {
		fmt.Println("failed make iofs source")
		return err
	}

	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		fmt.Println("failed to DOWN the migrations")
		return err
	}

	return nil
}
