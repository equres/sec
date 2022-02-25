package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/seccache"
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

		sc := seccache.NewSECCache(DB, S)
		switch args[0] {
		case "sitemap":
			if S.Verbose {
				log.Info("Generating a sitemap.xml file...")
			}
			err := GenerateSitemap(sc)
			if err != nil {
				return err
			}
		case "stats":
			if S.Verbose {
				log.Info("Generating & caching stats in redis...")
			}

			err := sc.GenerateStatsCache()
			if err != nil {
				return err
			}
		case "pages":
			if S.Verbose {
				log.Info("Generating & caching pages in redis...")
			}
			err := sc.GenerateHomePageDataCache()
			if err != nil {
				return err
			}
			if S.Verbose {
				log.Info("Generating & caching Months in Year page in redis...")
			}
			err = sc.GenerateMonthDayCIKDataCache()
			if err != nil {
				return err
			}

			if S.Verbose {
				log.Info("Generating & caching Companies page in redis...")
			}
			err = sc.GenerateCompanySlugsDataCache()
			if err != nil {
				return err
			}

			if S.Verbose {
				log.Info("Generating & caching SIC page in redis...")
			}
			err = sc.GenerateSICPageDataCache()
			if err != nil {
				return err
			}

			if S.Verbose {
				log.Info("Generating & caching Download Stats page in redis...")
			}
			err = sc.GenerateHourlyDownloadStatsPageDataCache()
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

func GenerateSitemap(sc *seccache.SECCache) error {
	sm := sitemap.New()

	var allURLs []string
	allURLs = append(allURLs, S.Config.Main.WebsiteURL)
	allURLs = append(allURLs, fmt.Sprintf("%vabout", S.Config.Main.WebsiteURL))
	allURLs = append(allURLs, fmt.Sprintf("%vcompany", S.Config.Main.WebsiteURL))

	yearMonthDayCIKURLs, err := sc.GenerateYearMonthDayCIKURLs(DB, S.Config.Main.WebsiteURL)
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
