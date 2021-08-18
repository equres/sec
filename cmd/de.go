// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"fmt"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// deCmd represents the de command
var deCmd = &cobra.Command{
	Use:   "de",
	Short: "toggle 'download enable' flag for statements from yyyy/mm month",
	Long:  `toggle 'download enable' flag for statements from yyyy/mm month`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := util.CheckMigration()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := errors.New("please enter a year and month (for example: 2021 or 2021/06)")
			return err
		}

		year_month := args[0]

		year, month, err := util.ParseYearMonth(year_month)
		if err != nil {
			return err
		}

		err = util.CheckRSSAvailability(year, month)
		if err != nil {
			return err
		}

		err = util.Downloadability(year, month, true)
		if err != nil {
			return err
		}

		fmt.Println("Successfully set download enabled for:", year_month)
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
