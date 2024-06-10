package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Log struct {
		Path string `yaml:"path"`
	} `yaml:"log"`
	DB struct {
		Dir string `yaml:"dir"`
	}
	Server struct {
		Host     string `yaml:"host"`
		SSHPort  int    `yaml:"ssh_port"`
		HTTPPort int    `yaml:"http_port"`
		CertsDir string `yaml:"certs_dir"`
	}
	Neynar struct {
		APIKey   string `yaml:"api_key"`
		ClientID string `yaml:"client_id"`
		BaseUrl  string `yaml:"base_url"`
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

func (c *Config) BaseURL() string {
	if c.Server.HTTPPort == 443 {
		return "https://" + c.Server.Host
	}
	return fmt.Sprintf("http://%s:%d", c.Server.Host, c.Server.HTTPPort)
}
