// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"os"
	"strconv"
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
		current_year, _, _ := time.Now().Date()
		year, err := strconv.Atoi(year_month[0:4])
		if err != nil {
			panic(err)
		}

		if year > current_year {
			err = errors.New("please enter a year equal to or below current year")
			panic(err)
		}

		var month int
		if len(year_month) > 4 {
			if string(year_month[4]) == "/" {
				month, err = strconv.Atoi(year_month[5:])
				if err != nil {
					panic(err)
				}
			}
		}

		db, err := util.ConnectDB()
		if err != nil {
			panic(err)
		}

		// If month = 0 then disable download for all months of that year ELSE then save only that specific month (e.g. 2021/05 so save only 05)
		if month == 0 {
			for i := 1; i <= 12; i++ {
				err = util.SaveWorklist(year, i, false, db)
				if err != nil {
					panic(err)
				}
			}
		} else {
			err = util.SaveWorklist(year, month, false, db)
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
