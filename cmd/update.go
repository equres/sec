// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/equres/sec/pkg/database"
	"github.com/spf13/cobra"
)

// Update/Insert data into DB

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "get critical RSS feeds and download them ~/.sec/data directory, parse them and update the DB",
	Long:  `get critical RSS feeds and download them ~/.sec/data directory, parse them and update the DB`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := S.TickerUpdateAll(DB)
		if err != nil {
			return err
		}

		return nil
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
