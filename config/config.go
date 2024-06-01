package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Neynar struct {
		APIKey   string `yaml:"api_key"`
		ClientID string `yaml:"client_id"`
		HubURL   string `yaml:"hub_url"`
	}
}

func ReadConfig(path string) (*Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal([]byte(data), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
