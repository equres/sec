package seccik

import "github.com/jmoiron/sqlx"

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
