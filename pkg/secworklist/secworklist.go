package secworklist

import "github.com/jmoiron/sqlx"

type Worklist struct {
	Year         int  `db:"year"`
	Month        int  `db:"month"`
	WillDownload bool `db:"will_download"`
}

func WillDownloadGet(db *sqlx.DB, isNewestToOlder bool) ([]Worklist, error) {
	// Retrieve from DB
	var worklist []Worklist

	query := "SELECT year, month, will_download FROM sec.worklist WHERE will_download = true ORDER BY year, month ASC"
	if isNewestToOlder {
		query = "SELECT year, month, will_download FROM sec.worklist WHERE will_download = true ORDER BY year DESC, month DESC"
	}
	err := db.Select(&worklist, query)
	if err != nil {
		return nil, err
	}
	return worklist, nil
}

func UniqueYears(db *sqlx.DB) ([]int, error) {
	// Retrieve from DB
	var worklistYears []int

	err := db.Select(&worklistYears, "SELECT DISTINCT year FROM sec.worklist WHERE will_download = true ORDER BY year DESC")
	if err != nil {
		return nil, err
	}
	return worklistYears, nil
}

func MonthsInYear(db *sqlx.DB, year int) ([]int, error) {
	// Retrieve from DB
	var worklistMonths []int

	err := db.Select(&worklistMonths, "SELECT month FROM sec.worklist WHERE will_download = true AND year = $1 ORDER BY year ASC", year)
	if err != nil {
		return nil, err
	}
	return worklistMonths, nil
}

func Save(year int, month int, willDownload bool, db *sqlx.DB) error {
	_, err := db.Exec(`
		INSERT INTO sec.worklist (year, month, will_download, created_at, updated_at) 
		VALUES ($1, $2, $3, NOW(), NOW()) 
		ON CONFLICT (month, year) 
		DO UPDATE SET
			will_download = EXCLUDED.will_download,
			updated_at=NOW() 
		WHERE 1=1
		AND worklist.year=EXCLUDED.year
		AND worklist.month=EXCLUDED.month ;`, year, month, willDownload)
	if err != nil {
		return err
	}
	return nil
}
