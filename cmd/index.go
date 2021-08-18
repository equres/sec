// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "saves XBRL File data in the database",
	Long:  `saves XBRL File data in the database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := util.ConnectDB()
		if err != nil {
			return err
		}

		worklist, err := util.WorklistWillDownloadGet(db)
		if err != nil {
			return err
		}

		sec, err := util.NewSEC("https://www.sec.gov")
		if err != nil {
			return err
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", sec.Config.Main.CacheDir, formatted)

			rssFile, err := sec.ParseRSSGoXML(fileURL)
			if err != nil {
				err = fmt.Errorf("you did not download any files yet. Run sec dow data to download the files, then run sec index to save their information to the database")
				return err
			}

			for _, v1 := range rssFile.Channel.Item {
				err = sec.SecItemFileUpsert(db, v1)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
