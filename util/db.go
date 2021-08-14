// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"embed"
	"fmt"

	"github.com/DavidHuie/gomigrate"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/johejo/golang-migrate-extra/source/iofs"
)

func ConnectDB() (*sqlx.DB, error) {
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

func MigrateUp(db *sqlx.DB) error {
	//go:embed _migrations
	var fileSystem embed.FS

	config, err := LoadConfig(".")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	d, err := iofs.New(fileSystem, "migrations")
	if err != nil {
		fmt.Println("failed to make migrations")
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DBURLString)
	if err != nil {
		fmt.Println("failed make iofs source")
		return nil
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		fmt.Println("failed to UP the migrations " + err.Error())
		return nil
	}

	return nil
}

func MigrateDown(db *sqlx.DB) error {
	migrator, _ := gomigrate.NewMigrator(db.DB, gomigrate.Postgres{}, "./migrations")

	err := migrator.Rollback()
	if err != nil {
		return err
	}

	return nil
}
