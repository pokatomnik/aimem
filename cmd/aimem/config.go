package main

import (
	"os"
	"path"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	// e.g. http://127.0.0.1:11434
	OllamaHost string `toml:"ollamaHost"`

	// e.g. "nomic-embed-text:latest"
	EmbeddingModel string `toml:"embeddingModel"`

	// How many results to get max
	TopK int `toml:"topK"`

	// Embedding dimensions
	Dimensions uint32 `toml:"dimensions"`
}

func (c *Config) GetDimensions() uint32 {
	return c.Dimensions
}

func (c *Config) GetOllamaHost() string {
	return c.OllamaHost
}

func (c *Config) GetEmbeddingModel() string {
	return c.EmbeddingModel
}

func (c *Config) GetTopK() int {
	return c.TopK
}

func InitConfig() (*Config, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := path.Join(userHome, ".config", "aimem", "config.toml")

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
