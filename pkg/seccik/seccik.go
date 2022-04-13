package seccik

import (
	"fmt"

	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/equres/sec/pkg/sec"
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
	err := db.Select(&companyNames, "SELECT title FROM sec.tickers WHERE cik = $1 AND title IS NOT NULL", cik)
	if err != nil {
		return "", err
	}

	if len(companyNames) < 1 {
		err = db.Select(&companyNames, "SELECT companyname FROM sec.secitemfile WHERE ciknumber = $1 AND companyname IS NOT NULL;", cik)
		if err != nil {
			return "", err
		}

		if len(companyNames) < 1 {
			return "", nil
		}
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

func GetCIKsFromTxtFile(s *sec.SEC, db *sqlx.DB) error {
	filePath := filepath.Join(s.Config.Main.CacheDir, "/Archives/edgar/cik-lookup-data.txt")

	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	txtFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer txtFile.Close()

	var existingCiks []int
	err = db.Select(&existingCiks, "SELECT DISTINCT cik FROM sec.ciks;")
	if err != nil {
		return err
	}
	existingCiksMap := make(map[int]bool)
	for _, v := range existingCiks {
		existingCiksMap[v] = true
	}

	match := regexp.MustCompile(`:([0-9]+):`)
	scanner := bufio.NewScanner(txtFile)
	for scanner.Scan() {
		cik, err := strconv.Atoi(strings.ReplaceAll(match.FindString(scanner.Text()), ":", ""))
		if err != nil {
			return err
		}

		if _, ok := existingCiksMap[cik]; ok {
			continue
		}

		err = SaveCIK(db, cik)
		if err != nil {
			return err
		}
	}

	return nil
}
