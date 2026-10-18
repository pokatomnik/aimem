package remember

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

type (
	embedder interface {
		GetEmbeddings(ctx context.Context, texts ...string) (ollama.Embeddings, error)
	}

	store interface {
		AddFact(ctx context.Context, vector []float32, content string) error
	}

	Remember struct {
		e embedder
		s store
	}
)

func New(e embedder, s store) *Remember {
	return &Remember{e, s}
}

func (r *Remember) Remember(ctx context.Context, facts ...string) error {
	embeddings, err := r.e.GetEmbeddings(ctx, facts...)
	if err != nil {
		return err
	}

	if len(embeddings) != len(facts) {
		return apperrors.ErrorEmbeddingsQuantityNotEqualToFactsQuantity
	}

	// TODO: think about batched calls
	for i := range embeddings {
		f := facts[i]
		e := embeddings[i]

		err = r.s.AddFact(ctx, e, f)

		if err != nil {
			return err
		}
	}

	return nil
}
