// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// dowIndexCmd represents the index command
var dowIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Download only index (RSS/XML) files into the local disk",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		db, err := database.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		s, err := sec.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		s.Verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		s.Debug, err = cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		if s.Verbose {
			fmt.Println("Checking/Downloading index files...")
		}

		err = s.DownloadIndex(db)
		if err != nil {
			return err
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
