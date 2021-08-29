// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
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
		var total_size_zip int

		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		downloader := download.NewDownloader(RootConfig)

		// For organizing the output
		tabWriter := tabwriter.NewWriter(os.Stdout, 12, 0, 2, ' ', 0)

		if S.Verbose {
			fmt.Fprint(tabWriter, "File Name", "\t\t", "Uncompressed Sized", "\t\t", "ZIP Sizes", "\n")
		}
		for _, v := range worklist {
			var file_size float64
			var file_size_zip int

			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			filePath := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", S.Config.Main.CacheDir, formatted)

			_, err = downloader.FileInCache(filePath)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download the necessary files then run sec dest again")
			}

			if S.Verbose {
				fmt.Fprint(tabWriter, fmt.Sprintf("xbrlrss-%v.xml", formatted), "\t\t")
			}

			rssFile, err := S.ParseRSSGoXML(filePath)
			if err != nil {
				return err
			}

			for _, item := range rssFile.Channel.Item {
				for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
					if xbrlFile.Size == "" {
						fmt.Printf("File %s size is ZERO!\n", xbrlFile.File)
						continue
					}

					val, err := strconv.ParseFloat(xbrlFile.Size, 64)
					if err != nil {
						return err
					}
					file_size += val
				}
			}
			if S.Verbose {
				fmt.Fprint(tabWriter, parseSize(file_size), "\t\t")
			}

			file_size_zip, err = S.CalculateRSSFilesZIP(rssFile)
			if err != nil {
				return err
			}

			if S.Verbose {
				fmt.Fprint(tabWriter, parseSize(float64(file_size_zip)), "\t\t", "\n")
			}

			total_size += file_size
			total_size_zip += file_size_zip
		}

		fmt.Fprint(tabWriter, "Total Size", "\t\t", parseSize(total_size), "\t\t", parseSize(float64(total_size_zip)))
		err = tabWriter.Flush()
		if err != nil {
			return err
		}
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
