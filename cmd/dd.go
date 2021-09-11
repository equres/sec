// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"fmt"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// ddCmd represents the dd command
var ddCmd = &cobra.Command{
	Use:   "dd",
	Short: "toggle 'download disable' flag for statements from yyyy/mm month ",
	Long:  `toggle 'download disable' flag for statements from yyyy/mm month `,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := errors.New("please enter a year or year/month (for example: 2021 or 2021/06)")
			return err
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

		err = S.Downloadability(DB, year, month, false)
		if err != nil {
			return err
		}

		fmt.Println("Successfully set download disabled for:", yearMonth)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
