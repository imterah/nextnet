package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Username     string `json:"username"`
	RefreshToken string `json:"token"`
	APIPath      string `json:"api_path"`
}

func ReadAndParseConfig(configFile string) (*Config, error) {
	configFileContents, err := os.ReadFile(configFile)

	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configFileContents, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}
