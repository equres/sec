// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// destCmd represents the dest command
var destCmd = &cobra.Command{
	Use:   "dest",
	Short: "Displaying disk space needed for all worklist that will be downloaded",
	Long:  `Displaying disk space needed for all worklist that will be downloaded`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var total_size float64

		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		err = S.DownloadIndex(DB)
		if err != nil {
			return err
		}

		for _, v := range worklist {
			var file_size float64

			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", S.Config.Main.CacheDir, formatted)

			if S.Verbose {
				fmt.Printf("Calculating space needed for file %v: ", fmt.Sprintf("xbrlrss-%v.xml", formatted))
			}

			rssFile, err := S.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			for _, item := range rssFile.Channel.Item {
				for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
					if xbrlFile.Size != "" {
						val, err := strconv.ParseFloat(xbrlFile.Size, 64)
						if err != nil {
							return err
						}
						file_size += val
					}
				}
			}
			if S.Verbose {
				fmt.Println(parseSize(file_size))
			}

			total_size += file_size
		}

		fmt.Printf("Size needed to download all files: %s\n", parseSize(total_size))
		return nil
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
