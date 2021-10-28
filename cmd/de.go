// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"errors"
	"log"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// deCmd represents the de command
var deCmd = &cobra.Command{
	Use:   "de",
	Short: "toggle 'download enable' flag for statements from yyyy/mm month",
	Long:  `toggle 'download enable' flag for statements from yyyy/mm month`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please enter a year and month (for example: 2021 or 2021/06)")
		}

		yearMonth := args[0]
		year, month, err := sec.ParseYearMonth(yearMonth)
		if err != nil {
			return err
		}

		err = sec.CheckRSSAvailability(year, month)
		if err != nil {
			return err
		}

		err = S.Downloadability(DB, year, month, true)
		if err != nil {
			return err
		}

		log.Print("Successfully set download enabled for:", yearMonth)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
