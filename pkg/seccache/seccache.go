package seccache

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

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

func (sc *SECCache) GenerateTopFiveRecentFilingsJSON() (string, error) {
	recentFilings, err := secutil.GetFiveRecentFilings(sc.DB)
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

	var eventStatsArrFormattedDates []secevent.EventStat
	for _, eventStat := range eventStatsArr {
		eventStat.Date = eventStat.DateTime.Format("2006-01-02")
		eventStatsArrFormattedDates = append(eventStatsArrFormattedDates, eventStat)
	}

	allStatsJSON, err := json.Marshal(eventStatsArrFormattedDates)
	if err != nil {
		return "", err
	}

	return string(allStatsJSON), nil
}

func (sc *SECCache) GenerateBackupStatsJSON() (string, error) {
	eventStatsArr, err := secevent.GetBackupEventStats(sc.DB)
	if err != nil {
		return "", err
	}

	var eventStatsArrFormattedDates []secevent.BackupEventStat
	for _, eventStat := range eventStatsArr {
		eventStat.Date = eventStat.DateTime.Format("2006-01-02")
		eventStatsArrFormattedDates = append(eventStatsArrFormattedDates, eventStat)
	}

	allStatsJSON, err := json.Marshal(eventStatsArrFormattedDates)
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

func (sc *SECCache) GenerateBackupStatsCache() error {
	statsJSON, err := sc.GenerateBackupStatsJSON()
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECBackupStats, statsJSON)
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
	formattedFilingsJSON, err := sc.GenerateTopFiveRecentFilingsJSON()
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

	var allCompanies []sec.Company
	for slug, company := range companySlugs {
		allCompanies = append(allCompanies, sec.Company{
			CompanyName: company.CompanyName,
			Slug:        slug,
		})
	}

	var companiesHTML string
	for index, company := range allCompanies {
		companiesHTML += fmt.Sprintf(`
		<tr>
			<td>%v</td>
			<td><a href="/company/%v">%v</a></td>
		</tr>`, index+1, company.Slug, company.CompanyName)
	}

	err = sc.S.Cache.MustSet(cache.SECCompanySlugsHTML, string(companiesHTML))
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

	type FormattedFiling struct {
		CompanyName string
		FillingDate string
		FormType    string
		FilingsURL  string
	}

	formattedFilings := make(map[string][]FormattedFiling)

	for year, secItemFiles := range filings {
		for _, item := range secItemFiles {
			filingsURL := fmt.Sprintf("/filings/%v/%v/%v/%v", item.FillingDate.Year(), int(item.FillingDate.Month()), item.FillingDate.Day(), item.CIKNumber)
			formType := secutil.GetFullFormType(item.FormType)
			formattedFilings[year] = append(formattedFilings[year], FormattedFiling{
				CompanyName: item.CompanyName,
				FillingDate: item.FillingDate.Format("2006-01-02"),
				FormType:    formType,
				FilingsURL:  filingsURL,
			})
		}
	}

	type YearFilings struct {
		Year    string
		Filings []FormattedFiling
	}

	var allYearsFilings []YearFilings
	for year, filings := range formattedFilings {
		allYearsFilings = append(allYearsFilings, YearFilings{
			Year:    year,
			Filings: filings,
		})
	}

	sort.Slice(allYearsFilings, func(i, j int) bool {
		return allYearsFilings[i].Year > allYearsFilings[j].Year
	})

	var HTMLList string
	for _, yearFilings := range allYearsFilings {
		HTMLList += fmt.Sprintf("<li>%v</li>", yearFilings.Year)
		for _, filing := range yearFilings.Filings {
			HTMLList += fmt.Sprintf(`
			<ul>
				<li><a href="%v">%v %v %v (filled: %v)</a></li>
			</ul>
		`, filing.FilingsURL, filing.CompanyName, yearFilings.Year, filing.FormType, filing.FillingDate)
		}
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECCompanyFilingsHTML, cik), HTMLList)
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

func (sc *SECCache) GenerateHourlyDownloadStatsPageDataCache() error {
	allDownloadStats, err := secevent.GetDownloadEventStatsByHour(sc.DB)
	if err != nil {
		return err
	}

	dates := make(map[string]string)
	for _, stat := range allDownloadStats {
		formattedDate, err := time.Parse(time.RFC3339, stat.Date)
		if err != nil {
			return err
		}

		dates[stat.Date] = formattedDate.Format("2006-01-02")
	}

	datesJSON, err := json.Marshal(dates)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECDownloadDates, string(datesJSON))
	if err != nil {
		return err
	}

	hourlyDownloadStats := make(map[string][]secevent.DownloadEventStatsByHour)
	for hour := 0; hour < 24; hour++ {
		for unformattedDate, formattedDate := range dates {
			var hourDateDownloadStats secevent.DownloadEventStatsByHour
			for _, stat := range allDownloadStats {
				if stat.Date == unformattedDate && stat.Hour == fmt.Sprint(hour) {
					hourDateDownloadStats = stat
				}
			}

			if hourDateDownloadStats.Hour == "" {
				hourDateDownloadStats.Hour = fmt.Sprint(hour)
				hourDateDownloadStats.Date = formattedDate
				hourDateDownloadStats.FilesDownloaded = 0
			}

			hourlyDownloadStats[formattedDate] = append(hourlyDownloadStats[formattedDate], hourDateDownloadStats)
		}
	}

	hourlyDownloadStatsJSON, err := json.Marshal(hourlyDownloadStats)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECHourlyDownloadStats, string(hourlyDownloadStatsJSON))
	if err != nil {
		return err
	}

	var hours []int
	for hour := 0; hour < 24; hour++ {
		hours = append(hours, hour)
	}

	hoursJSON, err := json.Marshal(hours)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(cache.SECHours, string(hoursJSON))
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) GenerateCompaniesDataCache() error {
	companies, err := sec.GetAllCompanies(sc.DB)
	if err != nil {
		return err
	}

	companyJSON, err := json.Marshal(companies)
	if err != nil {
		return err
	}

	err = sc.S.Cache.MustSet(fmt.Sprintf("%v", cache.SECCompanies), string(companyJSON))
	if err != nil {
		return err
	}

	return nil
}

func (sc *SECCache) AddToFilingsNotificationQueue(filing sec.SECItemFile) error {
	filingJSON, err := json.Marshal(filing)
	if err != nil {
		return err
	}

	err = sc.S.Cache.QueuePush(cache.SECFilingNotification, string(filingJSON))
	if err != nil {
		return err
	}

	return nil
}

// Take out of the FilingsNotificationQueue
func (sc *SECCache) GetFromFilingsNotificationQueue() (sec.SECItemFile, error) {
	filingJSON, err := sc.S.Cache.QueuePop(cache.SECFilingNotification)
	if err != nil {
		return sec.SECItemFile{}, err
	}

	var filing sec.SECItemFile
	err = json.Unmarshal([]byte(filingJSON), &filing)
	if err != nil {
		return sec.SECItemFile{}, err
	}

	return filing, nil
}
