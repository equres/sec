package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/equres/sec/pkg/server"
	"github.com/jmoiron/sqlx"
	"github.com/snabb/sitemap"
	"github.com/spf13/cobra"
)

// regenCmd represents the regen command
var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Generate a new sitemap for the website",
	Long:  `Generate a new sitemap for the website`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		out, err := os.Create("./_cache/sitemap.xml")
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = sm.WriteTo(out)
		if err != nil {
			return err
		}

		// Ping to Google Search Engine
		_, err = http.Get("https://www.google.com/ping?sitemap=https://equres.com/_cache/sitemap.xml")
		if err != nil {
			return err
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
