// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"github.com/equres/sec/pkg/database"
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
		return S.ForEachWorklist(DB, S.IndexZIPFileContent, "")
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
