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
	"github.com/equres/sec/pkg/seccik"
	"github.com/equres/sec/pkg/secdata"
	"github.com/equres/sec/pkg/secindex"
	"github.com/equres/sec/pkg/secticker"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	"github.com/spf13/cobra"
)

var GlobalWillIndexSECData bool

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

		err = secticker.UpdateAll(S, DB)
		if err != nil {
			return err
		}

		S.Log("Inserting SIC Code List...")
		err = secindex.IndexSICCodes(S, DB)
		if err != nil {
			return err
		}

		S.Log("Inserting CIK From Txt File...")
		err = seccik.GetCIKsFromTxtFile(S, DB)
		if err != nil {
			return err
		}

		if S.Config.IndexMode.FinancialStatementDataSets == "enabled" || S.Config.IndexMode.FinancialStatementDataSets == "true" || GlobalWillIndexSECData {
			S.Log("Indexing Financial Statement Data Sets...")
			secData := secdata.NewSECData(secdata.NewSECDataOpsFSDS())
			err = secData.IndexData(S, DB)
			if err != nil {
				return err
			}

			S.Log("Indexing Mutual Fund Data...")
			secData = secdata.NewSECData(secdata.NewSECDataOpsMFD())
			err = secData.IndexData(S, DB)
			if err != nil {
				return err
			}
		}

		worklist, err := secworklist.WillDownloadGet(DB, false)
		if err != nil {
			return err
		}

		var rssFiles []sec.RSSFile

		for _, v := range worklist {
			fileURL, err := secutil.FormatFilePathDate(S.Config.Main.CacheDir, v.Year, v.Month)
			if err != nil {
				return err
			}

			_, err = os.Stat(fileURL)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download all index files first")
			}

			rssFile, err := secutil.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			rssFiles = append(rssFiles, rssFile)
		}

		allFilesInRSS, err := secutil.MapFilesInWorklistGetAll(rssFiles)
		if err != nil {
			return err
		}

		filesInDB, err := secutil.MapFilesInDBGetAll(DB, S, allFilesInRSS)
		if err != nil {
			return err
		}

		totalCount := len(allFilesInRSS) - len(filesInDB)
		err = secindex.InsertAllSecItemFile(DB, S, rssFiles, filesInDB, totalCount)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	indexCmd.Flags().BoolVarP(&GlobalWillIndexSECData, "secdata", "a", false, "Index SEC data (e.g. financial statement data sets, mutual fund data...)")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
