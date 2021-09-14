// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
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
			return fmt.Errorf("please run sec dow index to download the necessary files")
		}

		// Check if ticker with exchange json files are in cache
		filePath = filepath.Join(S.Config.Main.CacheDir, "files/company_tickers_exchange.json")
		_, err = downloader.FileInCache(filePath)
		if err != nil {
			return fmt.Errorf("please run sec dow index to download the necessary files")
		}

		err = S.TickerUpdateAll(DB)
		if err != nil {
			return err
		}

		if S.Config.IndexMode.FinancialStatementDataSets == "enabled" || S.Config.IndexMode.FinancialStatementDataSets == "true" {
			err = S.IndexFinancialStatementDataSets(DB)
			if err != nil {
				return err
			}
		}

		err = S.ForEachWorklist(DB, S.InsertAllSecItemFile, "")
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
