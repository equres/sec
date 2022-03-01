// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/secutil"
	"github.com/equres/sec/pkg/secworklist"
	log "github.com/sirupsen/logrus"
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
		worklist, err := secworklist.WillDownloadGet(DB, false)
		if err != nil {
			return err
		}

		downloader := download.NewDownloader(RootConfig)
		downloader.IsEtag = true

		S.Log("File Name - Uncompressed Sized - ZIP Sizes")

		var totalSize float64
		var totalSizeZIP int
		totalSizeForYear := make(map[int]float64)
		totalSizeZIPForYear := make(map[int]int)
		for index, downloadable := range worklist {
			filePath, err := secutil.FormatFilePathDate(S.Config.Main.CacheDir, downloadable.Year, downloadable.Month)
			if err != nil {
				return err
			}

			_, err = downloader.FileInCache(filePath)
			if err != nil {
				return fmt.Errorf("please run sec dow index to download the necessary files then run sec dest again")
			}

			rssFile, err := secutil.ParseRSSGoXML(filePath)
			if err != nil {
				return err
			}

			var fileSize float64
			for _, item := range rssFile.Channel.Item {
				for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
					if xbrlFile.Size == "" {
						log.Info(fmt.Sprintf("File %s size is ZERO!\n", xbrlFile.File))
						continue
					}

					val, err := strconv.ParseFloat(xbrlFile.Size, 64)
					if err != nil {
						return err
					}
					fileSize += val
				}
			}

			fileSizeZIP, err := secutil.CalculateRSSFilesZIP(rssFile)
			if err != nil {
				return err
			}

			S.Log(fmt.Sprintf("fn %v %v %v", filepath.Base(filePath), parseSize(fileSize), parseSize(float64(fileSizeZIP))))

			if _, ok := totalSizeForYear[downloadable.Year]; !ok {
				totalSizeForYear[downloadable.Year] = 0
				totalSizeZIPForYear[downloadable.Year] = 0
			}

			totalSizeForYear[downloadable.Year] += fileSize
			totalSizeZIPForYear[downloadable.Year] += fileSizeZIP

			totalSize += fileSize
			totalSizeZIP += fileSizeZIP

			if downloadable.Month == 12 || len(worklist)-1 == index {
				S.Log(fmt.Sprintf("Year %v - %v - %v", downloadable.Year, parseSize(totalSizeForYear[downloadable.Year]), parseSize(float64(totalSizeZIPForYear[downloadable.Year]))))
			}
		}

		log.Info("Total Size", " - ", parseSize(totalSize), " - ", parseSize(float64(totalSizeZIP)), "\n")
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
