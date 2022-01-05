// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/secdow"
	"github.com/equres/sec/pkg/secutil"
	"github.com/spf13/cobra"
)

// dowDataCmd represents the data command
var dowDataCmd = &cobra.Command{
	Use:   "data",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		if S.Verbose {
			log.Info("Checking/Downloading index files...")
		}

		err := secdow.DownloadIndex(DB, S)
		if err != nil {
			return err
		}

		err = secutil.ForEachWorklist(S, DB, secdow.DownloadAllItemFiles, "Checking/Downloading XBRL files listed in index files...")
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	dowCmd.AddCommand(dowDataCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowDataCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowDataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
