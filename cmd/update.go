// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"os"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// Update/Insert data into DB

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "get critical RSS feeds and download them ~/.sec/data directory, parse them and update the DB",
	Long:  `get critical RSS feeds and download them ~/.sec/data directory, parse them and update the DB`,
	Run: func(cmd *cobra.Command, args []string) {
		sec := util.NewSEC("https://www.sec.gov/")

		db, err := util.ConnectDB()
		if err != nil {
			panic(err)
		}

		err = sec.TickerUpdateAll(db)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
