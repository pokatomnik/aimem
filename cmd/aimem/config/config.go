package config

import (
	_ "embed"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	// AI API
	OllamaHost         string `toml:"ollamaHost" comment:"Ollama Host, e.g. http://127.0.0.1:11434"`
	EmbeddingModel     string `toml:"embeddingModel" comment:"Ollama embedding model, do not change It to keep your database working, example: embeddinggemma-2:740m"`
	SummarizationModel string `toml:"summarizationModel" comment:"Ollama summarization model ID, example: gemma4:e2b"`

	// Search API
	TopK                int     `toml:"topK" comment:"How many search results to show, example: 5"`
	SimilarityThreshold float32 `toml:"similarityThreshold" comment:"The threshold beyond which a fact is considered relevant, example: 0.5. Min: 0, max: 1, greater or less values truncated"` // Embedding dimensions
	Dimensions          uint32  `toml:"dimensions" comment:"Vector database dimensions. Make sure your embedding model supports It"`

	// Public API
	MCPHTTPHost string `toml:"mcpHttpHost" comment:"MCP HTTP host to listen on. Default is 127.0.0.1:7777"`
	WebHTTPHost string `toml:"webHttpHost" comment:"Web search engine to listen on. Default is 127.0.0.1:8888"`

	// Database API
	Qdrant Qdrant `toml:"qdrant" comment:"Qdrant database configuration. Optional, will use embedded DB engine when not specified"`
}

func (c Config) String() string {
	toml, _ := toml.Marshal(c)

	return string(toml)
}

func DefaultConfig() Config {
	return Config{
		TopK:                5,
		SimilarityThreshold: 0.5,
		Dimensions:          768,
		OllamaHost:          "http://localhost:11434",
		EmbeddingModel:      "embeddinggemma-2:740m",
		SummarizationModel:  "gemma4:e2b",
		MCPHTTPHost:         "127.0.0.1:7777",
		WebHTTPHost:         "127.0.0.1:8888",
		Qdrant:              Qdrant{},
	}
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

func (c *Config) GetWebHTTPHost() string {
	return c.WebHTTPHost
}

func (c *Config) GetSimilarityThreshold() float32 {
	return max(0, min(1, c.SimilarityThreshold))
}

// Checks if required Qdrant fields are configured
// and Qdrant database can be used by the app
func (c *Config) IsQdrantConfigured() bool {
	return c.Qdrant.Host != ""
}

func (c *Config) GetQdrantHost() string {
	return c.Qdrant.Host
}

func (c *Config) GetQdrantAPIKey() string {
	return c.Qdrant.APIKey
}
