package ollama

import (
	"net/http"
	"strings"
)

type OllamaConfig interface {
	GetOllamaHost() string
	GetEmbeddingModel() string
	GetDimensions() uint32
}

type OllamaClient struct {
	config OllamaConfig
	client *http.Client
}

func New(client *http.Client, config OllamaConfig) *OllamaClient {
	return &OllamaClient{config, client}
}

func (o *OllamaClient) getHost() string {
	host := o.config.GetOllamaHost()
	for strings.HasSuffix(host, "/") {
		host = strings.TrimSuffix(host, "/")
	}

	return host
}
