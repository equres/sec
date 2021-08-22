// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/equres/sec/config"
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

	url := "https://www.sec.gov"
	fmt.Printf("URL [default: '%v']: ", url)
	err := AcceptInput(reader, &url)
	if err != nil {
		return err
	}

	db_user := "postgres"
	fmt.Println("Database Config:")
	fmt.Printf("User [default: '%v']: ", db_user)
	err = AcceptInput(reader, &db_user)
	if err != nil {
		return err
	}

	db_password := ""
	fmt.Printf("Password [default: '%v']: ", db_password)
	err = AcceptInput(reader, &db_password)
	if err != nil {
		return err
	}

	host := "localhost"
	fmt.Printf("Host [default: '%v']: ", host)
	err = AcceptInput(reader, &host)
	if err != nil {
		return err
	}

	db_name := "sec_project"
	fmt.Printf("DB Name [default: '%v']: ", db_name)
	err = AcceptInput(reader, &db_name)
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
		BaseURL:  url,
		CacheDir: "./cache",
	})

	cfg.SetDefault("database", config.DatabaseConfig{
		Driver:   "postgres",
		Host:     host,
		Port:     5432,
		Name:     db_name,
		Password: db_password,
		User:     db_user,
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
