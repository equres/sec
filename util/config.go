// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Main     MainConfig
}

type DatabaseConfig struct {
	DBDriver         string `mapstructure:"driver"`
	DBDataSourceName string `mapstructure:"data_source_name"`
	DBURLString      string `mapstructure:"database_string"`
}

type MainConfig struct {
	CacheDir string `mapstructure:"cache_dir"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
