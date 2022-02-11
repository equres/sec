package secsic

import (
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
