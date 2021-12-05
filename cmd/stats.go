// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/spf13/cobra"
)

// statsCmd represents the dow command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display the stats for the downloading",
	Long:  `Display the stats for the downloading`,
}

func init() {
	rootCmd.AddCommand(statsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
