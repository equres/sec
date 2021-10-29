// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/spf13/cobra"
)

// dlistCmd represents the dlist command
var dlistCmd = &cobra.Command{
	Use:   "dlist",
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
		log.Info("Year/Month that will be downloaded:")

		worklist, err := sec.WorklistWillDownloadGet(DB)
		if err != nil {
			return err
		}

		sort.SliceStable(worklist, func(i, j int) bool {
			return worklist[i].Year < worklist[j].Year
		})

		worklistMap := make(map[string]sec.Worklist)

		years := make(map[int]struct{})

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")
			worklistMap[formatted] = v
			years[v.Year] = struct{}{}
		}

		for k := range years {
			log.Info(k, " [")
			for i := 1; i <= 12; i++ {
				date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", k, i))
				if err != nil {
					return err
				}
				formatted := date.Format("2006-01")

				if _, ok := worklistMap[formatted]; ok && worklistMap[formatted].WillDownload {
					log.Info(fmt.Sprintf("+%d ", i))
					continue
				}
				log.Info(fmt.Sprintf("-%d ", i))
			}
			log.Info("]")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dlistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dlistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
