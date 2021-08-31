// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// indexzCmd represents the indexz command
var indexzCmd = &cobra.Command{
	Use:   "indexz",
	Short: "Loops through ZIP files and inserts in DB",
	Long:  `Loops through ZIP files and inserts in DB`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", S.Config.Main.CacheDir, formatted)

			rssFile, err := S.ParseRSSGoXML(fileURL)
			if err != nil {
				err = fmt.Errorf("you did not download any files yet. Run sec dow data to download the files, then run sec index to save their information to the database")
				return err
			}

			if S.Verbose {
				fmt.Printf("Inserting ZIP file data from xbrlrss-%v.xml\n", formatted)
			}

			totalCount := len(rssFile.Channel.Item)
			currentCount := 0
			for _, v1 := range rssFile.Channel.Item {
				zipPath := strings.ReplaceAll(v1.Enclosure.URL, S.BaseURL, "")

				zipCachePath := filepath.Join(RootConfig.Main.CacheDir, zipPath)
				_, err := os.Stat(zipCachePath)
				if err != nil {
					return fmt.Errorf("please run sec dowz to download all ZIP files then run sec indexz again to index them")
				}

				reader, err := zip.OpenReader(zipCachePath)
				if err != nil {
					return err
				}

				defer reader.Close()

				err = S.ZIPContentUpsert(DB, zipPath, reader.File)
				if err != nil {
					return err
				}

				if S.Verbose {
					fmt.Printf("[%d/%d] %s downloaded for current file...\n", currentCount, totalCount, time.Now().Format("2006-01-02 03:04:05"))
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexzCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexzCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexzCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
