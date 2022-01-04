// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/sec"
	"github.com/equres/sec/pkg/secextra"
	"github.com/equres/sec/pkg/secticker"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Download only index (RSS/XML) files into the local disk",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		downloader := download.NewDownloader(RootConfig)
		downloader.IsEtag = true
		downloader.Verbose = S.Verbose
		downloader.Debug = S.Debug

		// Check if ticker with no exchange json files are in cache
		filePath := filepath.Join(S.Config.Main.CacheDir, "files/company_tickers_exchange.json")
		_, err = downloader.FileInCache(filePath)
		if err != nil {
			log.Info("please run sec dow index to download the necessary files")
			return nil
		}

		// Check if ticker with exchange json files are in cache
		filePath = filepath.Join(S.Config.Main.CacheDir, "files/company_tickers_exchange.json")
		_, err = downloader.FileInCache(filePath)
		if err != nil {
			log.Info("please run sec dow index to download the necessary files")
			return nil
		}

		err = secticker.TickerUpdateAll(S, DB)
		if err != nil {
			return err
		}

		if S.Config.IndexMode.FinancialStatementDataSets == "enabled" || S.Config.IndexMode.FinancialStatementDataSets == "true" {
			err = secextra.IndexFinancialStatementDataSets(S, DB)
			if err != nil {
				return err
			}
		}

		worklist, err := secworklist.WillDownloadGet(DB)
		if err != nil {
			return err
		}

		var rssFiles []sec.RSSFile

		for _, v := range worklist {
			fileURL, err := S.FormatFilePathDate(S.Config.Main.CacheDir, v.Year, v.Month)
			if err != nil {
				return err
			}

			_, err = os.Stat(fileURL)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download all index files first")
			}

			rssFile, err := S.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			rssFiles = append(rssFiles, rssFile)
		}

		var totalCount int
		for _, rssFile := range rssFiles {
			for _, item := range rssFile.Channel.Item {
				totalCount += len(item.XbrlFiling.XbrlFiles.XbrlFile)
			}
		}

		err = S.InsertAllSecItemFile(DB, rssFiles, worklist, totalCount)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
