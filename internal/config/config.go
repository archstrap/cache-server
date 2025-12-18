package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	port               string
	host               string
	maxParallelization int
}

func (c *AppConfig) GetPort() string {
	return c.port
}

func (c *AppConfig) GetHost() string {
	return c.host
}

func (c *AppConfig) GetMaxParallelization() int {
	return c.maxParallelization
}

func (c *AppConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

func NewAppConfig() (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./resources")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &AppConfig{
		port:               viper.GetString("port"),
		host:               viper.GetString("host"),
		maxParallelization: viper.GetInt("maxParallelization"),
	}, nil
}
