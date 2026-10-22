package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

func (o *OllamaClient) LSEmbeddingModels(ctx context.Context) ([]ollama.OllamaModel, error) {
	baseHost := o.getHost()
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/api/tags", baseHost),
		nil,
	)
	if err != nil {
		return []ollama.OllamaModel{}, err
	}

	res, err := o.client.Do(req)
	if err != nil {
		return []ollama.OllamaModel{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []ollama.OllamaModel{}, apperrors.ErrCannotGetModels
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return []ollama.OllamaModel{}, err
	}

	var ollamaResponse OllamaModelsResponse

	err = json.Unmarshal(data, &ollamaResponse)
	if err != nil {
		return []ollama.OllamaModel{}, err
	}

	models := make([]ollama.OllamaModel, 0, len(ollamaResponse.Models))

	for _, m := range ollamaResponse.Models {
		capabilities := make([]ollama.ModelCapability, 0, len(m.Capabilities))
		hasEmbeddingsFeature := false
		for _, c := range m.Capabilities {
			if c == CapabilityEmbedding {
				hasEmbeddingsFeature = true
			}
			capabilities = append(capabilities, ollama.ModelCapability(c))
		}

		if !hasEmbeddingsFeature {
			continue
		}

		models = append(models, ollama.OllamaModel{
			Name:         m.Model,
			Model:        m.Model,
			Capabilities: capabilities,
		})
	}

	return models, nil
}
