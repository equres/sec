package seccache

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/equres/sec/pkg/cache"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/seccik"
	"github.com/equres/sec/pkg/secevent"
	"github.com/equres/sec/pkg/secextra"
	"github.com/equres/sec/pkg/secsic"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/equres/sec/pkg/server"
	"github.com/jmoiron/sqlx"
)

type SECCache struct {
	DB *sqlx.DB
	S  *sec.SEC
}

func NewSECCache(db *sqlx.DB, s *sec.SEC) *SECCache {
	return &SECCache{
		DB: db,
		S:  s,
	}
}

func GenerateTopFiveRecentFilingsJSON(db *sqlx.DB) (string, error) {
	recentFilings, err := secutil.GetFiveRecentFilings(db)
	if err != nil {
		return "", err
	}

	type FormattedFiling struct {
		CompanyName string
		FillingDate string
		FormType    string
		XbrlURL     string
	}

	var recentFilingsFormatted []FormattedFiling

	for _, filing := range recentFilings {
		filingURL := fmt.Sprintf("/filings/%v/%v/%v/%v", filing.FillingDate.Year(), int(filing.FillingDate.Month()), filing.FillingDate.Day(), filing.CIKNumber)

		formattedFiling := FormattedFiling{
			CompanyName: filing.CompanyName,
			FillingDate: filing.FillingDate.Format("2006-01-02"),
			FormType:    filing.FormType,
			XbrlURL:     filingURL,
		}
		recentFilingsFormatted = append(recentFilingsFormatted, formattedFiling)
	}

	recentFilingsData, err := json.Marshal(recentFilingsFormatted)
	if err != nil {
		return "", err
	}

	return string(recentFilingsData), nil
}

func (sc *SECCache) GenerateStatsJSON() (string, error) {
	eventStatsArr, err := secevent.GetEventStats(sc.DB)
	if err != nil {
		return "", err
	}
	allStats := make(map[string]int)
	for _, event := range eventStatsArr {
		statValue := 2
		if event.FilesBroken > 0 {
			statValue--
		}

		if event.FilesIndexed == 0 || event.FilesDownloaded == 0 {
			statValue--
		}

		allStats[event.Date] = statValue
	}

	allStatsJSON, err := json.Marshal(allStats)
	if err != nil {
		return "", err
	}

	return string(allStatsJSON), nil
}

func (sc *SECCache) GenerateStatsCache() error {
	statsJSON, err := sc.GenerateStatsJSON()
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECCacheStats, statsJSON)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateYearMonthDayCIKURLs(db *sqlx.DB, baseURL string) ([]string, error) {
	uniqueYears, err := secworklist.UniqueYears(sc.DB)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, year := range uniqueYears {
		urls = append(urls, fmt.Sprintf("%vfilings/%v", baseURL, year))

		months, err := secworklist.MonthsInYear(db, year)
		if err != nil {
			return nil, err
		}

		for _, month := range months {
			urls = append(urls, fmt.Sprintf("%vfilings/%v/%v", baseURL, year, month))

			days, err := secutil.GetFilingDaysFromMonthYear(sc.DB, year, month)
			if err != nil {
				return nil, err
			}
			for _, day := range days {
				urls = append(urls, fmt.Sprintf("%vfilings/%v/%v/%v", baseURL, year, month, day))

				companies, err := secutil.GetFilingCompaniesFromYearMonthDay(db, year, month, day)
				if err != nil {
					return nil, err
				}

				for _, company := range companies {
					urls = append(urls, fmt.Sprintf("%vfilings/%v/%v/%v/%v", baseURL, year, month, day, company.CIKNumber))
				}
			}
		}
	}

	return urls, nil
}

func (sc *SECCache) GenerateHomePageDataCache() error {
	formattedFilingsJSON, err := GenerateTopFiveRecentFilingsJSON(sc.DB)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECTopFiveRecentFilings, formattedFilingsJSON)
	if err != nil {
		return err
	}

	ciksCount, err := seccik.GetUniqueCIKCount(sc.DB)
	if err != nil {
		return err
	}
	err = sc.S.Cache.MustSet(cache.SECCIKsCount, ciksCount)
	if err != nil {
		return err
	}

	filesCount, err := secextra.GetUniqueFilesCount(sc.DB)
	if err != nil {
		return err
	}
	err = sc.S.Cache.MustSet(cache.SECFilesCount, filesCount)
	if err != nil {
		return err
	}

	companiesCount, err := secextra.GetUniqueFilesCompaniesCount(sc.DB)
	if err != nil {
		return err
	}
	err = sc.S.Cache.MustSet(cache.SECCompaniesCount, companiesCount)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateMonthDayCIKDataCache() error {
	years, err := secworklist.UniqueYears(sc.DB)
	if err != nil {
		return err
	}

	for _, year := range years {
		months, err := secworklist.MonthsInYear(sc.DB, year)
		if err != nil {
			return err
		}

		monthsJSON, err := json.Marshal(months)
		if err != nil {
			return err
		}

		err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECMonthsInYear, year), string(monthsJSON))
		if err != nil {
			return err
		}

		err = sc.GenerateDaysInMonthPageDataCache(year, months)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SECCache) GenerateDaysInMonthPageDataCache(year int, months []int) error {
	for _, month := range months {
		days, err := secutil.GetFilingDaysFromMonthYear(sc.DB, year, month)
		if err != nil {
			return err
		}

		daysJSON, err := json.Marshal(days)
		if err != nil {
			return err
		}

		err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v_%v", cache.SECDaysInMonth, year, month), string(daysJSON))
		if err != nil {
			return err
		}

		for _, day := range days {
			err = sc.GenerateCompaniesInDayPageData(year, month, day)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (sc *SECCache) GenerateCompaniesInDayPageData(year int, month int, day int) error {
	companies, err := secutil.GetFilingCompaniesFromYearMonthDay(sc.DB, year, month, day)
	if err != nil {
		return err
	}

	companiesJSON, err := json.Marshal(companies)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v_%v_%v", cache.SECCompaniesInDay, year, month, day), string(companiesJSON))
	if err != nil {
		return err
	}

	for _, company := range companies {
		cik, err := strconv.Atoi(company.CIKNumber)
		if err != nil {
			return err
		}
		err = sc.GenerateFilingsInDayPageDataCache(year, month, day, cik)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SECCache) GenerateFilingsInDayPageDataCache(year int, month int, day int, cik int) error {
	filings, err := secutil.SearchFilingsByYearMonthDayCIK(sc.DB, year, month, day, cik)
	if err != nil {
		return err
	}

	filingsJSON, err := json.Marshal(filings)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v_%v_%v_%v", cache.SECFilingsInDay, year, month, day, cik), string(filingsJSON))
	if err != nil {
		return err
	}

	err = sc.GenerateCompanyFilingsPageDataCache(cik)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateCompanySlugsDataCache() error {
	companies, err := sec.GetAllCompanies(sc.DB)
	if err != nil {
		return err
	}

	companySlugs := server.GetCompanySlugs(companies)

	companySlugsJSON, err := json.Marshal(companySlugs)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECCompanySlugs, string(companySlugsJSON))
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateCompanyFilingsPageDataCache(cik int) error {
	filings, err := sec.GetCompanyFilingsFromCIK(sc.DB, cik)
	if err != nil {
		return err
	}

	filingsJSON, err := json.Marshal(filings)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECCompanyFilings, cik), string(filingsJSON))
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateSICPageDataCache() error {
	sics, err := secsic.GetAllSICCodes(sc.DB)
	if err != nil {
		return err
	}

	sicsJSON, err := json.Marshal(sics)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECSICs, string(sicsJSON))
	if err != nil {
		return err
	}

	for _, sic := range sics {
		err = sc.GenerateCompaniesWithSICPageDataCache(sic.SIC)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SECCache) GenerateCompaniesWithSICPageDataCache(sic string) error {
	companies, err := secsic.GetAllCompaniesWithSIC(sc.DB, sic)
	if err != nil {
		return err
	}

	companiesJSON, err := json.Marshal(companies)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECCompaniesWithSIC, sic), string(companiesJSON))
	if err != nil {
		return err
	}

	return nil
}
