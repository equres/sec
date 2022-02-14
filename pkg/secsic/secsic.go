package secsic

import (
	"fmt"

	"github.com/equres/sec/pkg/sec"
	"github.com/jmoiron/sqlx"
)

func GetAllSICCodes(db *sqlx.DB) ([]sec.SIC, error) {
	var sics []sec.SIC
	err := db.Select(&sics, "SELECT sic, title FROM sec.sics;")
	if err != nil {
		return nil, err
	}

	return sics, nil
}

func GetAllCompaniesWithSIC(db *sqlx.DB, sic string) ([]sec.Company, error) {
	var companies []sec.Company
	err := db.Select(&companies, "SELECT DISTINCT companyname, ciknumber FROM sec.secitemfile WHERE assignedsic = $1 AND companyname IS NOT NULL;", sic)
	if err != nil {
		return nil, err
	}

	return companies, nil
}

func GetCategoryNameFromSIC(db *sqlx.DB, sic string) (string, error) {
	var categoryNames []string
	err := db.Select(&categoryNames, "SELECT title FROM sec.sics WHERE sic = $1;", sic)
	if err != nil {
		return "", err
	}

	if len(categoryNames) < 1 {
		return "", fmt.Errorf("Could not find the category name for this SIC Code: %v", sic)
	}

	return categoryNames[0], nil
}
