// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/equres/sec/pkg/config"
	"github.com/equres/sec/pkg/database"
	"github.com/equres/sec/pkg/sec"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var RootConfig config.Config
var defaultCfgPath string
var RateLimit time.Duration
var Verbose bool
var Debug bool
var DB *sqlx.DB
var S *sec.SEC

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
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	defaultCfgPath = filepath.Join(xdg.ConfigHome, "/.sec")

	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().BoolVar(&Verbose, "verbose", false, "Display the summarized version of progress")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Display additional details for debugging")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfgPath, fmt.Sprintf("config file (default is %v)", defaultCfgPath))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	var err error
	var cfg config.Config
	if cfgFile != defaultCfgPath {
		if _, err = os.Stat(cfgFile); err != nil {
			err = fmt.Errorf("file config '%v' was not found", cfgFile)
			cobra.CheckErr(err)
		}
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		filePath := filepath.Join(defaultCfgPath, "config.yaml")
		if _, err := os.Stat(filePath); err != nil {
			fmt.Println("you do not have a config file. Please create it by answering the questions below")
			err = GenerateConfig()
			if err != nil {
				cobra.CheckErr(err)
			}
			os.Exit(0)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		cobra.CheckErr(err)
	}
	RootConfig = cfg
	RateLimit, err = time.ParseDuration(fmt.Sprintf("%vms", RootConfig.Main.RateLimitMs))
	if err != nil {
		cobra.CheckErr(err)
	}

	DB, err = database.ConnectDB(RootConfig)
	if err != nil {
		cobra.CheckErr(err)
	}
	S, err = sec.NewSEC(RootConfig)
	if err != nil {
		cobra.CheckErr(err)
	}
	S.Verbose = Verbose
	S.Debug = Debug
}
