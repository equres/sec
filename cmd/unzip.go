// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"github.com/spf13/cobra"
)

// unzipCmd represents the unzip command
var unzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "extracts ZIP files to the cache unpacked directory",
	Long:  `extracts ZIP files to the cache unpacked directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return S.ForEachWorklist(DB, S.UnzipFiles, "")
	},
}

func init() {
	rootCmd.AddCommand(unzipCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unzipCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unzipCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
