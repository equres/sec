// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// destzCmd represents the destz command
var destzCmd = &cobra.Command{
	Use:   "destz",
	Short: "Displaying disk space needed for all worklist ZIPs that will be downloaded",
	Long:  `Displaying disk space needed for all worklist ZIPs that will be downloaded`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := util.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		sec, err := util.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			return err
		}

		err = sec.DownloadIndex()
		if err != nil {
			return err
		}

		var total_count float64
		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", sec.Config.Main.CacheDir, formatted)

			rssFile, err := sec.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			val, err := sec.CalculateRSSFilesZIP(rssFile)
			if err != nil {
				return err
			}
			total_count += val
		}

		fmt.Printf("Size needed to download all ZIP files: %s\n", parseSize(total_count))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(destzCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destzCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destzCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
