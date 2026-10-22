package config

import (
	"os"
	"path"

	toml "github.com/pelletier/go-toml/v2"
)

func InitConfig() (*Config, error) {
	configPath, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func configPath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(userHome, ".config", "aimem", "config.toml"), nil
}
