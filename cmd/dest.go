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

		s, err := sec.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		s.Verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		db, err := database.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		worklist, err := sec.WorklistWillDownloadGet(db)
		if err != nil {
			return err
		}

		err = s.DownloadIndex()
		if err != nil {
			return err
		}

		sizeChan := make(chan float64, len(worklist))
		errChan := make(chan error, 1)
		for _, v := range worklist {
			go CalculateSizeInRSSFile(s, v, sizeChan, errChan)
		}

		for i := 0; i < len(worklist); i++ {
			select {
			case size := <-sizeChan:
				total_size += size
			case err = <-errChan:
				return err
			}
		}

		fmt.Printf("Size needed to download all files: %s\n", parseSize(total_size))
		return nil
	},
}

func CalculateSizeInRSSFile(s *sec.SEC, worklist sec.Worklist, sizeChan chan float64, errChan chan error) {
	var file_size float64

	date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", worklist.Year, worklist.Month))
	if err != nil {
		errChan <- err
	}
	formatted := date.Format("2006-01")

	fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", s.Config.Main.CacheDir, formatted)

	if s.Verbose {
		fmt.Printf("Calculating space needed for file %v\n", fmt.Sprintf("xbrlrss-%v.xml", formatted))
	}

	rssFile, err := s.ParseRSSGoXML(fileURL)
	if err != nil {
		errChan <- err
	}

	for _, item := range rssFile.Channel.Item {
		for _, xbrlFile := range item.XbrlFiling.XbrlFiles.XbrlFile {
			val, err := strconv.ParseFloat(xbrlFile.Size, 64)
			if err != nil {
				errChan <- err
			}
			file_size += val
		}
	}
	sizeChan <- file_size
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
