package secevent

import (
	"encoding/json"
	"time"

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

type DownloadDayStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type IndexDayStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type EventStat struct {
	DateTime        time.Time `db:"events_date"`
	Date            string    `db:"date"`
	FilesDownloaded int       `db:"files_downloaded"`
	FilesBroken     int       `db:"files_broken"`
	FilesIndexed    int       `db:"files_indexed"`
}

type BackupEventStat struct {
	DateTime             time.Time `db:"events_date"`
	Date                 string    `db:"date"`
	SuccessfulFileBackup int       `db:"successful_file_backup"`
	FailedFileBackup     int       `db:"failed_file_backup"`
	SuccessfulDBBackup   int       `db:"successful_db_backup"`
	FailedDBBackup       int       `db:"failed_db_backup"`
}

type DownloadEventStatsByHour struct {
	Hour            string `db:"hour"`
	Date            string `db:"date"`
	FilesDownloaded int    `db:"files_downloaded"`
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

func GetEventStats(db *sqlx.DB) ([]EventStat, error) {
	var allEventStats []EventStat
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

func GetBackupEventStats(db *sqlx.DB) ([]BackupEventStat, error) {
	var allEventStats []BackupEventStat
	err := db.Select(&allEventStats, `
	SELECT 
		created_at::date as events_date, 
		COUNT(case when ev->>'job' = 'cache_compressed' AND ev->>'status' = 'success' then 1 end) as successful_file_backup,
		COUNT(case when ev->>'job' = 'cache_compressed' AND ev->>'status' = 'failed' then 1 end) as failed_file_backup,
		COUNT(case when ev->>'job' = 'db_backup' AND ev->>'status' = 'success' then 1 end) as successful_db_backup,
		COUNT(case when ev->>'job' = 'db_backup' AND ev->>'status' = 'failed' then 1 end) as failed_db_backup
	FROM sec.events GROUP BY created_at::date;
	`)
	if err != nil {
		return nil, err
	}

	return allEventStats, nil
}

func GetDownloadEventStatsByHour(db *sqlx.DB) ([]DownloadEventStatsByHour, error) {
	var allDownloadEventStats []DownloadEventStatsByHour
	err := db.Select(&allDownloadEventStats, `
	SELECT 
		EXTRACT(HOUR FROM created_at) as hour,
		created_at::date as date,
		COUNT(*) as files_downloaded
	FROM sec.events 
	WHERE 
		ev->>'event' = 'download' 
		AND ev->>'status' = 'success' 
	GROUP BY created_at::date, EXTRACT(HOUR FROM created_at);
	`)
	if err != nil {
		return nil, err
	}

	return allDownloadEventStats, nil
}

func GetLastSevenDaysDownloads(db *sqlx.DB) ([]DownloadDayStat, error) {
	var lastSevenDaysDownloadsCount []DownloadDayStat
	err := db.Select(&lastSevenDaysDownloadsCount, `
	SELECT
		COUNT(*) as count,
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'event' = 'download'
		AND ev->>'status' = 'success'
		AND created_at::date >= (current_date - 7)
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 7;
	`)
	if err != nil {
		return nil, err
	}

	return lastSevenDaysDownloadsCount, nil
}

func GetLastSevenDaysIndexes(db *sqlx.DB) ([]IndexDayStat, error) {
	var lastSevenDaysIndexesCount []IndexDayStat
	err := db.Select(&lastSevenDaysIndexesCount, `
	SELECT
		COUNT(*) as count,
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'event' = 'index'
		AND ev->>'status' = 'success'
		AND created_at::date >= (current_date - 7)
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 7;
	`)
	if err != nil {
		return nil, err
	}

	return lastSevenDaysIndexesCount, nil
}

func GetLastSuccessfulBackUpToCa2(db *sqlx.DB) (string, error) {
	var lastSuccessfulBackUpToCa2 []string
	err := db.Select(&lastSuccessfulBackUpToCa2, `
	SELECT
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'job' = 'ca2_rsync'
		AND ev->>'status' = 'success'
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 1;
	`)
	if err != nil {
		return "", err
	}

	if len(lastSuccessfulBackUpToCa2) == 0 {
		return "", nil
	}

	return lastSuccessfulBackUpToCa2[0], nil
}

func GetLastSuccessfulBackUpToWaw1(db *sqlx.DB) (string, error) {
	var lastSuccessfulBackUpToWaw1 []string
	err := db.Select(&lastSuccessfulBackUpToWaw1, `
	SELECT
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'job' = 'waw1_rsync'
		AND ev->>'status' = 'success'
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 1;
	`)
	if err != nil {
		return "", err
	}

	if len(lastSuccessfulBackUpToWaw1) == 0 {
		return "", nil
	}

	return lastSuccessfulBackUpToWaw1[0], nil
}

func GetLastSuccessfulDBBackup(db *sqlx.DB) (string, error) {
	var lastSuccessfulDBBackup []string
	err := db.Select(&lastSuccessfulDBBackup, `
	SELECT
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'job' = 'db_backup'
		AND ev->>'status' = 'success'
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 1;
	`)
	if err != nil {
		return "", err
	}

	if len(lastSuccessfulDBBackup) == 0 {
		return "", nil
	}

	return lastSuccessfulDBBackup[0], nil
}

func GetLastSuccessfulDBBackupToCa2(db *sqlx.DB) (string, error) {
	var lastSuccessfulDBBackupToCa2 []string
	err := db.Select(&lastSuccessfulDBBackupToCa2, `
	SELECT
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'job' = 'ca2_db_scp'
		AND ev->>'status' = 'success'
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 1;
	`)
	if err != nil {
		return "", err
	}

	if len(lastSuccessfulDBBackupToCa2) == 0 {
		return "", nil
	}

	return lastSuccessfulDBBackupToCa2[0], nil
}

func GetLastSuccessfulDBBackupToWaw1(db *sqlx.DB) (string, error) {
	var lastSuccessfulDBBackupToWaw1 []string
	err := db.Select(&lastSuccessfulDBBackupToWaw1, `
	SELECT
		created_at::date as date
	FROM sec.events
	WHERE
		ev->>'job' = 'waw1_db_scp'
		AND ev->>'status' = 'success'
	GROUP BY
		created_at::date
	ORDER BY
		created_at::date DESC
	LIMIT 1;
	`)
	if err != nil {
		return "", err
	}

	if len(lastSuccessfulDBBackupToWaw1) == 0 {
		return "", nil
	}

	return lastSuccessfulDBBackupToWaw1[0], nil
}
