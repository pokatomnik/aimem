package embeder

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/embeddings"
)

func (e *Embedder) GetEmbeddings(ctx context.Context, text string) (embeddings.Embeddings, error) {
	enc, err := e.getEncoder(ctx)
	if err != nil {
		return embeddings.Embeddings{}, err
	}

	vec, _, err := enc.EmbedText(text, "")
	vecs := [][]float32{vec}

	return embeddings.Embeddings(vecs), err
}
