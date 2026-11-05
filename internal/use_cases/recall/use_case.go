package recall

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/embeddings"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

type (
	embedder interface {
		GetEmbeddings(ctx context.Context, text string) (embeddings.Embeddings, error)
	}

	store interface {
		GetFactsList(ctx context.Context, userID string, vector []float32, topK int, similarityThreshold float32) (facts []fact.FactListView, err error)
	}

	config interface {
		GetTopK() int
		GetSimilarityThreshold() float32
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

func (r *Recall) Recall(ctx context.Context, userID string, question string) ([]fact.FactListView, error) {
	embeddings, err := r.e.GetEmbeddings(ctx, question)
	if err != nil {
		return []fact.FactListView{}, err
	}

	if len(embeddings) == 0 {
		return []fact.FactListView{}, apperrors.ErrEmptyEmbeddings
	}

	facts, err := r.s.GetFactsList(
		ctx,
		userID,
		[]float32(embeddings[0]),
		r.c.GetTopK(),
		r.c.GetSimilarityThreshold(),
	)
	if err != nil {
		return []fact.FactListView{}, err
	}

	return facts, nil
}
