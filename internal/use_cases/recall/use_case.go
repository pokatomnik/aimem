package recall

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/ollama"
)

type (
	embedder interface {
		GetEmbeddings(ctx context.Context, text string) (ollama.Embeddings, error)
	}

	store interface {
		GetFacts(ctx context.Context, vector []float32, topK int) (facts []fact.Fact, err error)
	}

	config interface {
		GetTopK() int
	}

	Recall struct {
		e embedder
		s store
		c config
	}
)

func New(e embedder, s store, c config) *Recall {
	return &Recall{e, s, c}
}

func (r *Recall) Recall(ctx context.Context, question string) ([]fact.Fact, error) {
	embeddings, err := r.e.GetEmbeddings(ctx, question)
	if err != nil {
		return []fact.Fact{}, err
	}

	if len(embeddings) == 0 {
		return []fact.Fact{}, apperrors.ErrorEmptyEmbeddings
	}

	facts, err := r.s.GetFacts(
		ctx,
		[]float32(embeddings[0]),
		r.c.GetTopK(),
	)
	if err != nil {
		return []fact.Fact{}, err
	}

	return facts, nil
}
