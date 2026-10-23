package config

import (
	_ "embed"
)

type Config struct {
	// e.g. http://127.0.0.1:11434
	OllamaHost string `toml:"ollamaHost"`

	// e.g. "nomic-embed-text:latest"
	EmbeddingModel string `toml:"embeddingModel"`

	// e.g. "gemma4:e2b"
	SummarizationModel string `toml:"summarizationModel"`

	// How many results to get max
	TopK int `toml:"topK"`

	// The threshold beyond which a fact is considered relevant
	// min: 0, max: 1
	SimilarityThreshold float32 `toml:"similarityThreshold"`

	// Embedding dimensions
	Dimensions uint32 `toml:"dimensions"`

	// e.g. http://127.0.0.1:8080
	MCPHTTPHost string `toml:"mcpHttpHost"`
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

func (c *Config) GetSummarizationModel() string {
	return c.SummarizationModel
}

func (c *Config) GetTopK() int {
	return c.TopK
}

func (c *Config) GetMCPHTTPHost() string {
	return c.MCPHTTPHost
}

func (c *Config) GetSimilarityThreshold() float32 {
	return max(0, min(1, c.SimilarityThreshold))
}
