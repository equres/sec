package cmd

import (
	"fmt"
	"os"

	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compares files in a directory to ones in index file",
	Long:  `Compares files in a directory to ones in index file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Info("Please type 'zip' to compare ZIP files, 'unzipped' to compare unzipped files, and 'raw' to compare the raw files")
			return nil
		}

		worklist, err := secworklist.WillDownloadGet(DB, false)
		if err != nil {
			return err
		}

		var rssFiles []sec.RSSFile
		var totalItems int
		var totalXbrlFiles int
		for _, v := range worklist {
			indexFilePath, err := secutil.FormatFilePathDate(S.Config.Main.CacheDir, v.Year, v.Month)
			if err != nil {
				return err
			}

			_, err = os.Stat(indexFilePath)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download all index files first")
			}

			rssFile, err := secutil.ParseRSSGoXML(indexFilePath)
			if err != nil {
				return err
			}

			rssFiles = append(rssFiles, rssFile)
			totalItems += len(rssFile.Channel.Item)
			for _, v1 := range rssFile.Channel.Item {
				totalXbrlFiles += len(v1.XbrlFiling.XbrlFiles.XbrlFile)
			}
		}

		switch args[0] {
		case "zip":
			return secutil.CompareZipFiles(S, DB, rssFiles, totalItems)
		case "unzipped":
			return secutil.CompareUnzippedFiles(S, DB, rssFiles, totalXbrlFiles)
		case "raw":
			return secutil.CompareRawFiles(S, DB, rssFiles, totalXbrlFiles)
		default:
			log.Info("Please type 'zip' to compare ZIP files, 'unzipped' to compare unzipped files, and 'raw' to compare the raw files")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compareCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compareCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
