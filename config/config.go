// Copyright (c) 2021 Equres LLC. All rights reserved.

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Main     MainConfig
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
	User     string `mapstructure:"user"`
}

type MainConfig struct {
	BaseURL  string `mapstructure:"baseurl"`
	CacheDir string `mapstructure:"cachedir"`
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

func (c *Config) DBGetURL() string {
	return fmt.Sprintf("%v://%v:%v@%v:%d/%v?sslmode=disable",
		c.Database.Driver,
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name)
}
