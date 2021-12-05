// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/equres/sec/pkg/database"
	"github.com/spf13/cobra"
)

var (
	GlobalEventInput  string
	GlobalJobInput    string
	GlobalStatusInput string
)

// eventCmd represents the dd command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "create an event in the events table ",
	Long:  `create an event in the events table `,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return database.CheckMigration(RootConfig)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return database.CreateOtherEvent(DB, GlobalEventInput, GlobalJobInput, GlobalStatusInput)
	},
}

func init() {
	rootCmd.AddCommand(eventCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	eventCmd.Flags().StringVar(&GlobalEventInput, "event", "cron", "event name")
	eventCmd.Flags().StringVar(&GlobalJobInput, "job", "cron", "job name")
	eventCmd.Flags().StringVar(&GlobalStatusInput, "status", "success", "status (success/failed)")
}
