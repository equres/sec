package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// DID NOT FIX IMPORT:
// ERROR EXPECTED

// dowzCmd represents the dowz command
var dowzCmd = &cobra.Command{
	Use:   "dowz",
	Short: "download all of the referenced file from XBRL index as ZIP files",
	Long:  `download all of the referenced file from XBRL index as ZIP files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		if S.Verbose {
			fmt.Println("Checking/Downloading index files...")
		}
		err = S.DownloadIndex(DB)
		if err != nil {
			return err
		}

		rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", S.Config.Main.RateLimitMs))
		if err != nil {
			return err
		}

		downloader := download.NewDownloader(S.Config)
		downloader.Verbose = S.Verbose
		downloader.Debug = S.Debug

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", S.Config.Main.CacheDir, formatted)

			rssFile, err := S.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			var current_count int
			total_count := len(rssFile.Channel.Item)
			for _, v1 := range rssFile.Channel.Item {
				if v1.Enclosure.URL != "" {

					not_download, err := downloader.FileCorrect(DB, v1.Enclosure.URL)
					if err != nil {
						return err
					}

					if !not_download {
						err = downloader.DownloadFile(DB, v1.Enclosure.URL)
						if err != nil {
							return err
						}
						time.Sleep(rateLimit)
					}

					current_count++
					if !S.Verbose {
						fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", current_count, total_count, (total_count - current_count))
					}

					if S.Verbose {
						fmt.Printf("[%d/%d] %s downloaded...\n", current_count, total_count, time.Now().Format("2006-01-02 03:04:05"))
					}
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dowzCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowzCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowzCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
