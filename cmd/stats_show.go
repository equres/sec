// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/cache"
	"github.com/spf13/cobra"
)

// statsShowCmd represents the data command
var statsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the download stats",
	Long:  `Show the download stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pool := cache.CreateRedisPool()

		stats, err := cache.GetAllStats(pool)
		if err != nil {
			return err
		}

		log.Info("Year/Month - Day - DownloadOK - DownloadFailed")
		for _, stat := range stats {
			log.Info(stat.YearMonth, " - ", stat.Day, " - ", stat.DownloadOK, " - ", stat.DownloadFailed)
		}
		return nil
	},
}

func init() {
	statsCmd.AddCommand(statsShowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statsShowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statsShowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
