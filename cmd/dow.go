// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// dowCmd represents the dow command
var dowCmd = &cobra.Command{
	Use:   "dow",
	Short: "Download all files in the downloadable years",
	Long:  `Download all files in the downloadable years`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := errors.New("please type 'index' to download only RSS files and type 'data' to download actual reports (e.g. sec dow index)")
			return err
		}

		if args[0] != "index" && args[0] != "data" {
			err := errors.New("please type 'index' to download only RSS files and type 'data' to download actual reports (e.g. sec dow index)")
			return err
		}

		db, err := util.ConnectDB()
		if err != nil {
			return err
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			return err
		}

		sec := util.NewSEC("https://www.sec.gov")
		sec.Verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		config, err := util.LoadConfig(".")
		if err != nil {
			return err
		}

		err = sec.DownloadIndex()
		if err != nil {
			return err
		}

		// Get Count of Items in RSSFile
		var total_count int
		var current_count int
		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			filepath := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", config.Database.CacheDir, formatted)
			rssFile, err := sec.ParseRSSGoXML(filepath)
			if err != nil {
				return err
			}

			for _, v1 := range rssFile.Channel.Item {
				total_count += len(v1.XbrlFiling.XbrlFiles.XbrlFile)
			}
		}

		if args[0] == "data" {
			for _, v := range worklist {
				date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
				if err != nil {
					return err
				}
				formatted := date.Format("2006-01")
				fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", config.Database.CacheDir, formatted)

				rssFile, err := sec.ParseRSSGoXML(fileURL)
				if err != nil {
					return err
				}

				for _, v1 := range rssFile.Channel.Item {
					for _, v2 := range v1.XbrlFiling.XbrlFiles.XbrlFile {
						err = sec.DownloadFile(v2.URL, config)
						if err != nil {
							return err
						}
						current_count++
						if !sec.Verbose {
							fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", current_count, total_count, (total_count - current_count))
						}

						if sec.Verbose {
							fmt.Printf("[%d/%d] %s downloaded...\n", current_count, total_count, time.Now().Format("2006-01-02 03:04:05"))
						}
						time.Sleep(1 * time.Second)
					}

					err = util.SaveSecItemFile(db, v1)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dowCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
