// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Download only index (RSS/XML) files into the local disk",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		for _, v := range worklist {
			fileURL, err := S.FormatFilePathDate(S.Config.Main.CacheDir, v.Year, v.Month)
			if err != nil {
				return err
			}

			rssFile, err := S.ParseRSSGoXML(fileURL)
			if err != nil {
				err = fmt.Errorf("you did not download any files yet. Run sec dow data to download the files, then run sec index to save their information to the database")
				return err
			}

			if S.Verbose {
				fmt.Printf("Insert data from %v\n", filepath.Base(fileURL))
			}

			for _, v1 := range rssFile.Channel.Item {
				err = S.SecItemFileUpsert(DB, v1)
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
