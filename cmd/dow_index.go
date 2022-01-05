// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/equres/sec/pkg/secdow"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// dowIndexCmd represents the index command
var dowIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Download only index (RSS/XML) files into the local disk",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := secdow.DownloadTickerFile(DB, S, "files/company_tickers.json")
		if err != nil {
			return err
		}

		err = secdow.DownloadTickerFile(DB, S, "files/company_tickers_exchange.json")
		if err != nil {
			return err
		}

		if S.Verbose {
			log.Info("Checking/Downloading index files...")
		}

		err = secdow.DownloadIndex(DB, S)
		if err != nil {
			return err
		}

		if RootConfig.IndexMode.FinancialStatementDataSets == "enabled" || RootConfig.IndexMode.FinancialStatementDataSets == "true" {
			if S.Verbose {
				log.Info("Downloading financial statement data sets...:")
			}

			err = secdow.DownloadFinancialStatementDataSets(DB, S)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	dowCmd.AddCommand(dowIndexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowIndexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowIndexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
