// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/equres/sec/pkg/secdata"
	"github.com/equres/sec/pkg/secdow"

	"github.com/spf13/cobra"
)

var GlobalWillDownloadSECData bool

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

		S.Log("Checking/Downloading index files...")

		err = secdow.DownloadIndex(DB, S)
		if err != nil {
			return err
		}

		if GlobalWillDownloadSECData {
			S.Log("Downloading financial statement data sets...:")

			secData := secdata.NewSECData(secdata.NewSECDataOpsFSDS())
			err = secData.DownloadSECData(DB, S)
			if err != nil {
				return err
			}

			S.Log("Downloading mutual fund data...:")

			secData = secdata.NewSECData(secdata.NewSECDataOpsMFD())
			err = secData.DownloadSECData(DB, S)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	dowCmd.AddCommand(dowIndexCmd)

	dowIndexCmd.Flags().BoolVarP(&GlobalWillDownloadSECData, "secdata", "a", false, "Download SEC data (e.g. financial statement data sets, mutual fund data...)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowIndexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowIndexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
