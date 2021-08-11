/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
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
			panic(err)
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			panic(err)
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				panic(err)
			}
			formatted := date.Format("2006-01")

			rssFile, err := sec.ParseRSSGoXML("Archives/edgar/monthly/xbrlrss-" + formatted + ".xml")
			if err != nil {
				panic(err)
			}

			for _, item := range rssFile.Channel.Item {
				for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
					val, err := strconv.ParseFloat(xbrlFile.Size, 64)
					if err != nil {
						panic(err)
					}
					size += val
				}
			}
		}

		fmt.Printf("Size needed to download all files: %s", parseSize(size))
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
