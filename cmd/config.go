// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/equres/sec/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "setup the config file for the SEC program",
	Long:  `setup the config file for the SEC program`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return GenerateConfig()
	},
}

func GenerateConfig() error {
	reader := bufio.NewReader(os.Stdin)

	user, err := user.Current()
	if err != nil {
		return err
	}

	url := "https://www.sec.gov"
	fmt.Printf("URL [default: '%v']: ", url)
	err = AcceptInput(reader, &url)
	if err != nil {
		return err
	}

	port := ":8000"
	fmt.Printf("Server Port [default: '%v']: ", port)
	err = AcceptInput(reader, &port)
	if err != nil {
		return err
	}

	rateLimit := "100"
	fmt.Printf("Rate Limit [default: '%v' Milliseconds]: ", rateLimit)
	err = AcceptInput(reader, &rateLimit)
	if err != nil {
		return err
	}

	retrylimit := "3"
	fmt.Printf("Rate Limit [default: '%v']: ", retrylimit)
	err = AcceptInput(reader, &retrylimit)
	if err != nil {
		return err
	}

	dbUser := user.Username
	fmt.Println("Database Config:")
	fmt.Printf("User [default: '%v']: ", dbUser)
	err = AcceptInput(reader, &dbUser)
	if err != nil {
		return err
	}

	dbPassword := ""
	fmt.Printf("Password [default: '%v']: ", dbPassword)
	err = AcceptInput(reader, &dbPassword)
	if err != nil {
		return err
	}

	host := "localhost"
	fmt.Printf("Host [default: '%v']: ", host)
	err = AcceptInput(reader, &host)
	if err != nil {
		return err
	}

	dbName := user.Username
	fmt.Printf("DB Name [default: '%v']: ", dbName)
	err = AcceptInput(reader, &dbName)
	if err != nil {
		return err
	}

	fmt.Println("Index Mode Config:")
	financialStatementDataSets := "enabled"
	fmt.Printf("Financial Statement Data Sets [default: '%v'] (enabled/disabled): ", financialStatementDataSets)
	err = AcceptInput(reader, &financialStatementDataSets)
	if err != nil {
		return err
	}

	companyfacts := "enabled"
	fmt.Printf("Company Facts [default: '%v'] (enabled/disabled): ", companyfacts)
	err = AcceptInput(reader, &companyfacts)
	if err != nil {
		return err
	}

	cfg := viper.New()

	if _, err = os.Stat(cfgFile); err != nil {
		err = os.MkdirAll(cfgFile, 0755)
		if err != nil {
			return err
		}
	}

	if _, err = os.Stat(filepath.Join(cfgFile, "config.yaml")); err != nil {
		_, err = os.Create(filepath.Join(cfgFile, "config.yaml"))
		if err != nil {
			return err
		}
	}

	cfg.AddConfigPath(cfgFile)
	cfg.SetConfigType("yaml")
	cfg.SetConfigName("config")

	cfg.SetDefault("main", config.MainConfig{
		BaseURL:          url,
		CacheDir:         "./cache",
		RateLimitMs:      rateLimit,
		RetryLimit:       retrylimit,
		CacheDirUnpacked: "./unzipped_cache",
		ServerPort:       port,
	})

	cfg.SetDefault("indexmode", config.IndexModeConfig{
		FinancialStatementDataSets: financialStatementDataSets,
		CompanyFacts:               companyfacts,
	})

	cfg.SetDefault("database", config.DatabaseConfig{
		Driver:   "postgres",
		Host:     host,
		Port:     5432,
		Name:     dbName,
		Password: dbPassword,
		User:     dbUser,
	})

	err = cfg.WriteConfig()
	if err != nil {
		return err
	}
	err = cfg.ReadInConfig()
	if err != nil {
		return err
	}

	fmt.Println("Successfully created the config")

	return nil
}

func AcceptInput(reader *bufio.Reader, data *string) error {
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	val := strings.TrimSuffix(input, "\n")
	val = strings.TrimSpace(val)
	if val != "" {
		*data = val
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
