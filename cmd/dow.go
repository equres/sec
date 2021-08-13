// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"os"
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
			fmt.Println(err)
			os.Exit(1)
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		sec := util.NewSEC("https://www.sec.gov")

		config, err := util.LoadConfig(".")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", sec.BaseURL, formatted)
			err = sec.DownloadFile(fileURL, config)
			if err != nil {
				panic(err)
			}
			time.Sleep(1 * time.Second)
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			formatted := date.Format("2006-01")
			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", config.CacheDir, formatted)

			rssFile, err := sec.ParseRSSGoXML(fileURL)
			if err != nil {
				panic(err)
			}

			for _, v1 := range rssFile.Channel.Item {
				for _, v2 := range v1.XbrlFiling.XbrlFiles.XbrlFile {
					err = sec.DownloadFile(v2.URL, config)
					if err != nil {
						panic(err)
					}
					time.Sleep(1 * time.Second)
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(dowCmd)

	dowCmd.PersistentFlags().Bool("verbose", false, "Display the summarized version of progress")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
