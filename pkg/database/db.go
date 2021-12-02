// Copyright (c) 2021 Koszek Systems. All rights reserved.

package database

import (
	"embed"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/johejo/golang-migrate-extra/source/iofs"
)

type IndexEvent struct {
	Event  string `json:"event"`
	File   string `json:"file"`
	Status string `json:"status"`
}

type DownloadEvent struct {
	Event  string `json:"event"`
	File   string `json:"file"`
	Status string `json:"status"`
}

type OtherEvent struct {
	Event  string `json:"event"`
	Job    string `json:"job"`
	Status string `json:"status"`
}

func ConnectDB(config config.Config) (*sqlx.DB, error) {
	// Connect to DB
	db, err := sqlx.Open(config.Database.Driver, config.DBGetURL())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateUp(db *sqlx.DB, fs embed.FS, config config.Config) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Info("failed to make migrations")
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DBGetURL())
	if err != nil {
		log.Info("failed make iofs source")
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Info("failed to UP the migrations")
		return err
	}

	return nil
}

func MigrateDown(db *sqlx.DB, fs embed.FS, config config.Config) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Info("failed to make migrations")
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, config.DBGetURL())
	if err != nil {
		log.Info("failed make iofs source")
		return err
	}

	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		log.Info("failed to DOWN the migrations")
		return err
	}

	return nil
}

func CheckMigration(config config.Config) error {
	db, err := ConnectDB(config)
	if err != nil {
		return err
	}

	// Check if migrated
	_, err = db.Exec("SELECT 'sec.tickers'::regclass")
	if err != nil {
		log.Info("looks like you're running sec for the first time. Please initialize the database with sec migrate up")
		return err
	}
	return nil
}

func CreateIndexEvent(db *sqlx.DB, file string, status string) error {
	event := IndexEvent{
		Event:  "index",
		File:   file,
		Status: status,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		return err
	}

	return nil
}

func CreateDownloadEvent(db *sqlx.DB, file string, status string) error {
	event := DownloadEvent{
		Event:  "download",
		File:   file,
		Status: status,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		return err
	}

	return nil
}

func CreateOtherEvent(db *sqlx.DB, eventName string, job string, status string) error {
	event := OtherEvent{
		Event:  eventName,
		Job:    job,
		Status: status,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		return err
	}

	return nil
}

func SkipFileInsert(db *sqlx.DB, fullurl string) error {
	_, err := db.Exec(`INSERT INTO sec.skipped_files (url) VALUES ($1)`, fullurl)
	if err != nil {
		return err
	}

	return nil
}

func IsSkippedFile(db *sqlx.DB, fullurl string) (bool, error) {
	var skippedFiles []string
	err := db.Select(&skippedFiles, "SELECT url FROM sec.skipped_files WHERE url = $1", fullurl)
	if err != nil {
		return true, err
	}

	if len(skippedFiles) > 0 {
		return true, err
	}

	return false, nil
}
