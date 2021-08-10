// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"os"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// ddCmd represents the dd command
var ddCmd = &cobra.Command{
	Use:   "dd",
	Short: "toggle 'download disable' flag for statements from yyyy/mm month ",
	Long:  `toggle 'download disable' flag for statements from yyyy/mm month `,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := errors.New("please enter a year and month (for example: 2021 or 2021/06)")
			panic(err)
		}

		year_month := args[0]

		var month int
		var year int

		switch len(year_month) {
		case 4:
			date, err := time.Parse("2006", year_month)
			if err != nil {
				panic(err)
			}
			year = date.Year()
		case 6:
			date, err := time.Parse("2006/1", year_month)
			if err != nil {
				panic(err)
			}
			year = date.Year()
			month = int(date.Month())
		case 7:
			date, err := time.Parse("2006/01", year_month)
			if err != nil {
				panic(err)
			}
			year = date.Year()
			month = int(date.Month())
		default:
			err := errors.New("please enter a valid date ('2021' or '2021/05')")
			panic(err)
		}

		db, err := util.ConnectDB()
		if err != nil {
			panic(err)
		}

		if month != 0 {
			err = util.SaveWorklist(year, month, false, db)
			if err != nil {
				panic(err)
			}
			return
		}

		for i := 1; i <= 12; i++ {
			err = util.SaveWorklist(year, i, false, db)
			if err != nil {
				panic(err)
			}
		}

		os.Exit(0)
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
