// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

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
		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		downloader := download.NewDownloader(RootConfig)
		downloader.IsEtag = true

		// For organizing the output
		tabWriter := tabwriter.NewWriter(os.Stdout, 12, 0, 2, ' ', 0)

		if S.Verbose {
			fmt.Fprint(tabWriter, "File Name", "\t", "Uncompressed Sized", "\t", "ZIP Sizes", "\n")
		}

		var totalSize float64
		var totalSizeZIP int
		for _, v := range worklist {
			filePath, err := S.FormatFilePathDate(S.Config.Main.CacheDir, v.Year, v.Month)
			if err != nil {
				return err
			}

			_, err = downloader.FileInCache(filePath)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download the necessary files then run sec dest again")
			}

			if S.Verbose {
				fmt.Fprint(tabWriter, fmt.Sprintf("%v", filepath.Base(filePath)), "\t\t")
				err = tabWriter.Flush()
				if err != nil {
					return err
				}
			}

			rssFile, err := S.ParseRSSGoXML(filePath)
			if err != nil {
				return err
			}

			var fileSize float64
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
					fileSize += val
				}
			}
			if S.Verbose {
				fmt.Fprint(tabWriter, parseSize(fileSize), "\t\t")
			}

			fileSizeZIP, err := S.CalculateRSSFilesZIP(rssFile)
			if err != nil {
				return err
			}

			if S.Verbose {
				fmt.Fprint(tabWriter, parseSize(float64(fileSizeZIP)), "\t\t", "\n")
			}

			err = tabWriter.Flush()
			if err != nil {
				return err
			}

			totalSize += fileSize
			totalSizeZIP += fileSizeZIP
		}

		fmt.Fprint(tabWriter, "Total Size", "\t\t", parseSize(totalSize), "\t\t", parseSize(float64(totalSizeZIP)), "\n")
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
