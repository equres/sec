// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// dowCmd represents the dow command
var dowCmd = &cobra.Command{
	Use:   "dow",
	Short: "Download all files in the downloadable years",
	Long:  `Download all files in the downloadable years`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := util.ConnectDB()
		if err != nil {
			panic(err)
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			panic(err)
		}

		sec := util.NewSEC("https://sec.gov/")
		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				panic(err)
			}
			formatted := date.Format("2006-01")
			fileURL := fmt.Sprintf("Archives/edgar/monthly/xbrlrss-%v.xml", formatted)

			rssFile, err := sec.ParseRSSGoXML(fileURL)
			if err != nil {
				panic(err)
			}

			err = sec.DownloadXbrlFiles(rssFile, fileURL)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(dowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
