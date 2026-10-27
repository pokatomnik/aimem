package config

import (
	"os"
	"path"

	toml "github.com/pelletier/go-toml/v2"
)

// Reads configuration from a file
// located ~/.config/aimem/config.toml,
// returns error if file is missing
func GetConfig() (*Config, error) {
	configPath, err := ConfigPath()
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

// Get a user-local configuration file path
func ConfigPath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(userHome, ".config", "aimem", "config.toml"), nil
}
