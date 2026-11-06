package embeder

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/embeddings"
)

func (e *Embedder) GetEmbeddings(_ context.Context, text string) (embeddings.Embeddings, error) {
	enc, err := e.getEncoder()
	if err != nil {
		e.logger.Error("get encoder failed", "text_length", len(text), "error", err)

		return embeddings.Embeddings{}, err
	}

	vec, err := enc.EmbedText(text)
	if err != nil {
		e.logger.Error("embed text failed", "text_length", len(text), "error", err)

		return nil, err
	}

	vecs := [][]float32{vec}

	return embeddings.Embeddings(vecs), err
}
