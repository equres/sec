// Copyright (c) 2021 Equres LLC. All rights reserved.

package util

import "github.com/spf13/viper"

type Config struct {
	DBDriver         string `mapstructure:"DB_DRIVER"`
	DBDataSourceName string `mapstructure:"DB_DATA_SOURCE_NAME"`
	CacheDir         string `mapstructure:"CACHE_DIR"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
