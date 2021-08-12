// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"github.com/DavidHuie/gomigrate"
	"github.com/jmoiron/sqlx"
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
	migrator, _ := gomigrate.NewMigrator(db.DB, gomigrate.Postgres{}, "./migrations")

	err := migrator.Migrate()
	if err != nil {
		return err
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
