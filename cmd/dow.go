// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"github.com/spf13/cobra"
)

// dowCmd represents the dow command
var dowCmd = &cobra.Command{
	Use:   "dow",
	Short: "Download files in the downloadable worklist years",
	Long:  `Download files in the downloadable worklist years`,
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
