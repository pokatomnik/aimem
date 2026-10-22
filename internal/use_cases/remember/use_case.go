package remember

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

type (
	ai interface {
		GetEmbeddings(ctx context.Context, text string) (ollama.Embeddings, error)
		Summarize(ctx context.Context, text string) (string, error)
	}

	store interface {
		AddFact(ctx context.Context, vector []float32, summary, content string) error
	}

	Remember struct {
		ai ai
		st store
	}
)

func New(e ai, s store) *Remember {
	return &Remember{e, s}
}

func (r *Remember) Remember(ctx context.Context, fact string) (string, error) {
	summary, embeddings, err := r.getEmbeddingsAndSummary(ctx, fact)
	if err != nil {
		return "", err
	}

	if len(embeddings) != 1 {
		return "", apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity
	}

	embedding := embeddings[0]
	err = r.st.AddFact(ctx, embedding, summary, fact)

	if err != nil {
		return "", err
	}

	return summary, nil
}
