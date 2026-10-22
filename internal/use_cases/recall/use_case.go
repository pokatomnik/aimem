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
		GetFactsList(ctx context.Context, vector []float32, topK int) (facts []fact.FactListView, err error)
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

func (r *Recall) Recall(ctx context.Context, question string) ([]fact.FactListView, error) {
	embeddings, err := r.e.GetEmbeddings(ctx, question)
	if err != nil {
		return []fact.FactListView{}, err
	}

	if len(embeddings) == 0 {
		return []fact.FactListView{}, apperrors.ErrEmptyEmbeddings
	}

	facts, err := r.s.GetFactsList(
		ctx,
		[]float32(embeddings[0]),
		r.c.GetTopK(),
	)
	if err != nil {
		return []fact.FactListView{}, err
	}

	return facts, nil
}
