package secextra

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func GetUniqueFilesCount(db *sqlx.DB) (string, error) {
	var filesCount []string

	err := db.Select(&filesCount, "SELECT COUNT(DISTINCT XbrlFile) FROM sec.secItemFile;")
	if err != nil {
		return "", err
	}
	if len(filesCount) < 1 {
		return "", fmt.Errorf("could_not_get_count_of_unique_files")
	}

	return filesCount[0], nil
}

func GetUniqueFilesCompaniesCount(db *sqlx.DB) (string, error) {
	var companiesCount []string

	err := db.Select(&companiesCount, "SELECT COUNT(DISTINCT companyname) FROM sec.secItemFile;")
	if err != nil {
		return "", err
	}
	if len(companiesCount) < 1 {
		return "", fmt.Errorf("could_not_get_count_of_unique_companies")
	}

	return companiesCount[0], nil
}
