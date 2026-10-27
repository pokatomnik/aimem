package config

type Qdrant struct {
	Host   string `toml:"host" comment:"Qdrant host, example: http://127.0.0.1:6333"`
	APIKey string `toml:"apiKey" comment:"Qdrant API key, can be empty when no authorization required"`
}
