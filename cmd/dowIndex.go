// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// dowIndexCmd represents the index command
var dowIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Download only index (RSS/XML) files into the local disk",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		sec, err := util.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		sec.Verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		err = sec.DownloadIndex()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	dowCmd.AddCommand(dowIndexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowIndexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowIndexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
