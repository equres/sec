package seccik

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func SaveCIK(db *sqlx.DB, cik int) error {
	_, err := db.Exec(`
		INSERT INTO sec.ciks (cik, created_at, updated_at) 
		VALUES ($1,NOW(), NOW()) 
		ON CONFLICT (cik) 
		DO NOTHING;`, cik)
	if err != nil {
		return err
	}
	return nil
}

func GetCompanyNameFromCIK(db *sqlx.DB, cik int) (string, error) {
	var companyNames []string
	err := db.Select(&companyNames, "SELECT title FROM sec.tickers WHERE cik = $1", cik)
	if err != nil {
		return "", err
	}

	if len(companyNames) < 1 {
		return "", nil
	}

	return companyNames[0], nil
}

func GetUniqueCIKCount(db *sqlx.DB) (string, error) {
	var cikCount []string

	err := db.Select(&cikCount, "SELECT COUNT(DISTINCT cikNumber) FROM sec.secItemFile;")
	if err != nil {
		return "", err
	}
	if len(cikCount) < 1 {
		return "", fmt.Errorf("could_not_get_count_of_unique_ciks")
	}

	return cikCount[0], nil
}
