package main

import "github.com/caarlos0/env"

type Config struct {
	OllamaHost          string  `env:"AIMEM_OLLAMA_HOST" envDefault:"http://127.0.0.1:11434"`
	EmbeddingModel      string  `env:"AIMEM_EMBEDDING_MODEL,required"`
	SummarizationModel  string  `env:"AIMEM_SUMMARIZATION_MODEL,required"`
	TopK                int     `env:"AIMEM_TOP_K" envDefault:"5"`
	SimilarityThreshold float32 `env:"AIMEM_SIMILARITY_THRESHOLD" envDefault:"0.7"`
	Dimensions          uint    `env:"AIMEM_DIMENSIONS,required"`
	WebHTTPHost         string  `env:"AIMEM_WEB_HTTP_HOST" envDefault:"127.0.0.1:8888"`
	QdrantHost          string  `env:"AIMEM_QDRANT_HOST"`
	QdrantAPIKey        string  `env:"AIMEM_QDRANT_API_KEY"`
}

func GetConfig() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetDimensions() uint32 {
	return uint32(c.Dimensions)
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

func (c *Config) GetWebHTTPHost() string {
	return c.WebHTTPHost
}

func (c *Config) GetSimilarityThreshold() float32 {
	return max(0, min(1, c.SimilarityThreshold))
}

func (c *Config) RequireAuth() bool {
	return c.QdrantHost != ""
}

func (c *Config) GetQdrantHost() string {
	return c.QdrantHost
}

func (c *Config) GetQdrantAPIKey() string {
	return c.QdrantAPIKey
}
