// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// destCmd represents the dest command
var destCmd = &cobra.Command{
	Use:   "dest",
	Short: "Displaying disk space needed for all worklist that will be downloaded",
	Long:  `Displaying disk space needed for all worklist that will be downloaded`,
	Run: func(cmd *cobra.Command, args []string) {
		var size float64

		sec := util.NewSEC("https://sec.gov/")
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

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("Archives/edgar/monthly/xbrlrss-%v.xml", formatted)

			rssFile, err := sec.ParseRSSGoXML(fileURL)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, item := range rssFile.Channel.Item {
				for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
					val, err := strconv.ParseFloat(xbrlFile.Size, 64)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					size += val
				}
			}
		}

		fmt.Printf("Size needed to download all files: %s\n", parseSize(size))
	},
}

func parseSize(size float64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%g B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "kMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(destCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
