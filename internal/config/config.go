package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	port               string
	host               string
	maxParallelization int
}

type configFile struct {
	Port               string `yaml:"port"`
	Host               string `yaml:"host"`
	MaxParallelization int    `yaml:"maxParallelization"`
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
	port := Port
	if port == "" {
		port = c.port
	}
	return fmt.Sprintf("%s:%s", c.host, port)
}

func NewAppConfig() (*AppConfig, error) {
	data, err := os.ReadFile("./resources/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &AppConfig{
		port:               cfg.Port,
		host:               cfg.Host,
		maxParallelization: cfg.MaxParallelization,
	}, nil
}
