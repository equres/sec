// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

var cfgFile string
var RootConfig util.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sec",
	Short: "A simple but powerful search engine for SEC stock filings",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := util.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		RootConfig = config
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	dowCmd.PersistentFlags().Bool("verbose", false, "Display the summarized version of progress")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./.sec", "config file (default is ./.sec/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
