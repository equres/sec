package secevent

import (
	"encoding/json"

	"github.com/equres/sec/pkg/sec"
	"github.com/jmoiron/sqlx"
)

type IndexEvent struct {
	Event  string `json:"event"`
	File   string `json:"file"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type DownloadEvent struct {
	Event  string `json:"event"`
	File   string `json:"file"`
	URL    string `json:"url"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type UnzipEvent struct {
	Event  string `json:"event"`
	File   string `json:"file"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type OtherEvent struct {
	Event  string `json:"event"`
	Job    string `json:"job"`
	Status string `json:"status"`
}

func CreateIndexEvent(db *sqlx.DB, file string, status string, reason string) {
	event := IndexEvent{
		Event:  "index",
		File:   file,
		Status: status,
		Reason: reason,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		panic(err)
	}
}

func CreateDownloadEvent(db *sqlx.DB, file string, url string, status string, reason string) {
	event := DownloadEvent{
		Event:  "download",
		File:   file,
		URL:    url,
		Status: status,
		Reason: reason,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		panic(err)
	}
}

func CreateUnzipEvent(db *sqlx.DB, file string, status string, reason string) {
	event := UnzipEvent{
		Event:  "unzip",
		File:   file,
		Status: status,
		Reason: reason,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		panic(err)
	}
}

func CreateOtherEvent(db *sqlx.DB, eventName string, job string, status string) {
	event := OtherEvent{
		Event:  eventName,
		Job:    job,
		Status: status,
	}
	eventJson, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`INSERT INTO sec.events (ev) VALUES ($1)`, eventJson)
	if err != nil {
		panic(err)
	}
}

func GetEventStats(db *sqlx.DB) ([]sec.EventStat, error) {
	var allEventStats []sec.EventStat
	err := db.Select(&allEventStats, `
	SELECT 
		created_at::date as events_date, 
		COUNT(case when ev->>'event' = 'download' AND ev->>'status' = 'success' then 1 end) as files_downloaded,
		COUNT(case when (ev->>'event' = 'download' OR ev->>'event' = 'unzip') AND ev->>'status' = 'failed' then 1 end) as files_broken,
		COUNT(case when ev->>'event' = 'index' AND ev->>'status' = 'success' then 1 end) as files_indexed
	FROM sec.events GROUP BY created_at::date;
	`)
	if err != nil {
		return nil, err
	}

	return allEventStats, nil
}
