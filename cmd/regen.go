package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	log "github.com/sirupsen/logrus"
	"github.com/snabb/sitemap"

	"github.com/spf13/cobra"
)

// regenCmd represents the regen command
var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Generate a new sitemap for the website",
	Long:  `Generate a new sitemap for the website`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Info("please type 'sitemap' to generate the sitemap,'stats' to generate the stats, and 'pages' to generate the data for the web pages (e.g. sec regen sitemap)")
			return nil
		}

		switch args[0] {
		case "sitemap":
			if S.Verbose {
				log.Info("Generating a sitemap.xml file...")
			}
			err := GenerateSitemap()
			if err != nil {
				return err
			}
		case "stats":
			if S.Verbose {
				log.Info("Generating & caching stats in redis...")
			}
			statsJSON, err := GenerateStatsJSON(DB, S)
			if err != nil {
				return err
			}

			err = S.Cache.MustSet(cache.SECCacheStats, statsJSON)
			if err != nil {
				return err
			}
		case "pages":
			if S.Verbose {
				log.Info("Generating & caching pages in redis...")
			}
			err := GenerateHomePageDataCache(DB)
			if err != nil {
				return err
			}
			if S.Verbose {
				log.Info("Generating & caching Months in Year page in redis...")
			}
			err = GenerateMonthDayCIKDataCache(DB)
			if err != nil {
				return err
			}

			if S.Verbose {
				log.Info("Generating & caching Companies page in redis...")
			}
			err = GenerateCompanySlugsDataCache(DB)
			if err != nil {
				return err
			}

			if S.Verbose {
				log.Info("Generating & caching SIC page in redis...")
			}
			err = GenerateSICPageDataCache(DB)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("please type 'sitemap' to generate the sitemap and 'stats' to generate the stats (e.g. sec regen sitemap)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(regenCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// regenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// regenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func GenerateYearMonthDayCIKURLs(db *sqlx.DB, baseURL string) ([]string, error) {
	uniqueYears, err := secworklist.UniqueYears(DB)
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

			days, err := secutil.GetFilingDaysFromMonthYear(DB, year, month)
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

func GenerateCompanyPageURLs(db *sqlx.DB, baseURL string) ([]string, error) {
	companies, err := sec.GetAllCompanies(db)
	if err != nil {
		return nil, err
	}
	companySlugs := server.GetCompanySlugs(companies)

	var urls []string
	for slug := range companySlugs {
		urls = append(urls, fmt.Sprintf("%vcompany/%v", baseURL, slug))
	}

	return urls, nil
}

func GenerateSitemap() error {
	sm := sitemap.New()

	var allURLs []string
	allURLs = append(allURLs, S.Config.Main.WebsiteURL)
	allURLs = append(allURLs, fmt.Sprintf("%vabout", S.Config.Main.WebsiteURL))
	allURLs = append(allURLs, fmt.Sprintf("%vcompany", S.Config.Main.WebsiteURL))

	yearMonthDayCIKURLs, err := GenerateYearMonthDayCIKURLs(DB, S.Config.Main.WebsiteURL)
	if err != nil {
		return err
	}
	allURLs = append(allURLs, yearMonthDayCIKURLs...)

	companyPageURLs, err := GenerateCompanyPageURLs(DB, S.Config.Main.WebsiteURL)
	if err != nil {
		return err
	}
	allURLs = append(allURLs, companyPageURLs...)

	currentTime := time.Now().UTC()
	for _, URL := range allURLs {
		sm.Add(&sitemap.URL{
			Loc:        URL,
			LastMod:    &currentTime,
			ChangeFreq: sitemap.Daily,
		})
	}

	sitemap, err := os.Create(filepath.Join(S.Config.Main.CacheDir, "sitemap.xml"))
	if err != nil {
		return err
	}
	defer sitemap.Close()

	_, err = sm.WriteTo(sitemap)
	if err != nil {
		return err
	}

	// Ping to Google Search Engine
	_, err = http.Get("https://www.google.com/ping?sitemap=https://equres.com/_cache/sitemap.xml")
	if err != nil {
		return err
	}

	return nil
}

func GenerateStatsJSON(db *sqlx.DB, s *sec.SEC) (string, error) {
	eventStatsArr, err := secevent.GetEventStats(db)
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

func GenerateHomePageDataCache(db *sqlx.DB) error {
	formattedFilingsJSON, err := GenerateTopFiveRecentFilingsJSON(DB)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(cache.SECTopFiveRecentFilings, formattedFilingsJSON)
	if err != nil {
		return err
	}

	ciksCount, err := seccik.GetUniqueCIKCount(DB)
	if err != nil {
		return err
	}
	err = S.Cache.MustSet(cache.SECCIKsCount, ciksCount)
	if err != nil {
		return err
	}

	filesCount, err := secextra.GetUniqueFilesCount(DB)
	if err != nil {
		return err
	}
	err = S.Cache.MustSet(cache.SECFilesCount, filesCount)
	if err != nil {
		return err
	}

	companiesCount, err := secextra.GetUniqueFilesCompaniesCount(DB)
	if err != nil {
		return err
	}
	err = S.Cache.MustSet(cache.SECCompaniesCount, companiesCount)
	if err != nil {
		return err
	}

	return nil
}

func GenerateMonthDayCIKDataCache(db *sqlx.DB) error {
	years, err := secworklist.UniqueYears(db)
	if err != nil {
		return err
	}

	for _, year := range years {
		months, err := secworklist.MonthsInYear(db, year)
		if err != nil {
			return err
		}

		monthsJSON, err := json.Marshal(months)
		if err != nil {
			return err
		}

		err = S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECMonthsInYear, year), string(monthsJSON))
		if err != nil {
			return err
		}

		err = GenerateDaysInMonthPageDataCache(db, year, months)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateDaysInMonthPageDataCache(db *sqlx.DB, year int, months []int) error {
	for _, month := range months {
		days, err := secutil.GetFilingDaysFromMonthYear(db, year, month)
		if err != nil {
			return err
		}

		daysJSON, err := json.Marshal(days)
		if err != nil {
			return err
		}

		err = S.Cache.MustSet(fmt.Sprintf("%v_%v_%v", cache.SECDaysInMonth, year, month), string(daysJSON))
		if err != nil {
			return err
		}

		for _, day := range days {
			err = GenerateCompaniesInDayPageData(db, year, month, day)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GenerateCompaniesInDayPageData(db *sqlx.DB, year int, month int, day int) error {
	companies, err := secutil.GetFilingCompaniesFromYearMonthDay(db, year, month, day)
	if err != nil {
		return err
	}

	companiesJSON, err := json.Marshal(companies)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(fmt.Sprintf("%v_%v_%v_%v", cache.SECCompaniesInDay, year, month, day), string(companiesJSON))
	if err != nil {
		return err
	}

	for _, company := range companies {
		cik, err := strconv.Atoi(company.CIKNumber)
		if err != nil {
			return err
		}
		err = GenerateFilingsInDayPageDataCache(db, year, month, day, cik)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateFilingsInDayPageDataCache(db *sqlx.DB, year int, month int, day int, cik int) error {
	filings, err := secutil.SearchFilingsByYearMonthDayCIK(db, year, month, day, cik)
	if err != nil {
		return err
	}

	filingsJSON, err := json.Marshal(filings)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(fmt.Sprintf("%v_%v_%v_%v_%v", cache.SECFilingsInDay, year, month, day, cik), string(filingsJSON))
	if err != nil {
		return err
	}

	err = GenerateCompanyFilingsPageDataCache(db, cik)
	if err != nil {
		return err
	}

	return nil
}

func GenerateCompanySlugsDataCache(db *sqlx.DB) error {
	companies, err := sec.GetAllCompanies(db)
	if err != nil {
		return err
	}

	companySlugs := server.GetCompanySlugs(companies)

	companySlugsJSON, err := json.Marshal(companySlugs)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(cache.SECCompanySlugs, string(companySlugsJSON))
	if err != nil {
		return err
	}

	return nil
}

func GenerateCompanyFilingsPageDataCache(db *sqlx.DB, cik int) error {
	filings, err := sec.GetCompanyFilingsFromCIK(db, cik)
	if err != nil {
		return err
	}

	filingsJSON, err := json.Marshal(filings)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECCompanyFilings, cik), string(filingsJSON))
	if err != nil {
		return err
	}

	return nil
}

func GenerateSICPageDataCache(db *sqlx.DB) error {
	sics, err := secsic.GetAllSICCodes(db)
	if err != nil {
		return err
	}

	sicsJSON, err := json.Marshal(sics)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(cache.SECSICs, string(sicsJSON))
	if err != nil {
		return err
	}

	for _, sic := range sics {
		err = GenerateCompaniesWithSICPageDataCache(db, sic.SIC)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateCompaniesWithSICPageDataCache(db *sqlx.DB, sic string) error {
	companies, err := secsic.GetAllCompaniesWithSIC(db, sic)
	if err != nil {
		return err
	}

	companiesJSON, err := json.Marshal(companies)
	if err != nil {
		return err
	}

	err = S.Cache.MustSet(fmt.Sprintf("%v_%v", cache.SECCompaniesWithSIC, sic), string(companiesJSON))
	if err != nil {
		return err
	}

	return nil
}
