// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// dowDataCmd represents the data command
var dowDataCmd = &cobra.Command{
	Use:   "data",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		worklist, err := sec.WorklistWillDownloadGet(db)
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

		if s.Verbose {
			fmt.Println("Checking/Downloading index files...")
		}
		err = s.DownloadIndex()
		if err != nil {
			return err
		}

		// Get Count of Items in RSSFile
		var total_count int
		var current_count int

		if s.Verbose {
			fmt.Print("Calculating number of XBRL Files in the index files: ")
		}

		total_count, err = s.TotalXbrlFileCountGet(worklist, s.Config.Main.CacheDir)
		if err != nil {
			return err
		}

		if s.Verbose {
			fmt.Println(total_count)
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", s.Config.Main.CacheDir, formatted)

			rssFile, err := s.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			if s.Verbose {
				fmt.Println("Checking/Downloading XBRL files listed in index files...")
			}

			for _, v1 := range rssFile.Channel.Item {
				err = s.DownloadXbrlFileContent(v1.XbrlFiling.XbrlFiles.XbrlFile, s.Config, &current_count, total_count)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	dowCmd.AddCommand(dowDataCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowDataCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowDataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
