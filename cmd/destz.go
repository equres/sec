// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// destzCmd represents the destz command
var destzCmd = &cobra.Command{
	Use:   "destz",
	Short: "Displaying disk space needed for all worklist ZIPs that will be downloaded",
	Long:  `Displaying disk space needed for all worklist ZIPs that will be downloaded`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		s, err := sec.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		s.Verbose, err = cmd.Flags().GetBool("verbose")
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

		var total_size int
		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", s.Config.Main.CacheDir, formatted)

			if s.Verbose {
				fmt.Printf("Calculating space needed for file %v: ", fmt.Sprintf("xbrlrss-%v.xml", formatted))
			}

			rssFile, err := s.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			val, err := s.CalculateRSSFilesZIP(rssFile)
			if err != nil {
				return err
			}

			if s.Verbose {
				fmt.Println(parseSize(float64(val)))
			}

			total_size += val
		}

		fmt.Printf("Size needed to download all ZIP files: %s\n", parseSize(float64(total_size)))

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
