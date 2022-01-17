package cmd

import (
	"github.com/equres/sec/pkg/download"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh essential files in the program",
	Long:  `Refresh essential files in the program`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileURLs := []string{
			"https://www.sec.gov/corpfin/division-of-corporation-finance-standard-industrial-classification-sic-code-list",
		}

		downloader := download.NewDownloader(S.Config)
		downloader.IsEtag = true
		downloader.Verbose = S.Verbose
		downloader.Debug = S.Debug
		downloader.TotalDownloadsCount = len(fileURLs)
		downloader.CurrentDownloadCount = 0

		for _, URL := range fileURLs {
			err := downloader.DownloadFile(DB, URL)
			if err != nil {
				log.Info("error_downloading_file ", URL)
				continue
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
