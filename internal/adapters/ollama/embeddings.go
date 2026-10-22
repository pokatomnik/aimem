package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

func (o *OllamaClient) GetEmbeddings(ctx context.Context, text string) (ollama.Embeddings, error) {
	baseHost := o.getHost()
	oer := OllamaEmbedRequest{
		Model:      o.config.GetEmbeddingModel(),
		Input:      []string{text},
		Dimensions: o.config.GetDimensions(),
	}

	oerBytes, err := json.Marshal(oer)
	if err != nil {
		return ollama.Embeddings{}, err
	}

	br := bytes.NewReader(oerBytes)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/api/embed", baseHost),
		br,
	)
	if err != nil {
		return ollama.Embeddings{}, err
	}

	res, err := o.client.Do(req)
	if err != nil {
		return ollama.Embeddings{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ollama.Embeddings{}, apperrors.ErrGetEmbeddings
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return ollama.Embeddings{}, err
	}

	var response OllamaEmbedResponse

	err = json.Unmarshal(data, &response)
	if err != nil {
		return ollama.Embeddings{}, err
	}

	return ollama.Embeddings(response.Embeddings), nil
}
