package recall

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/embeddings"
	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/shared/logging"
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
		e      embedder
		s      store
		c      config
		logger logging.Logger
	}
)

func New(e embedder, s store, c config, logger logging.Logger) *Recall {
	return &Recall{e: e, s: s, c: c, logger: logger}
}

func (r *Recall) Recall(ctx context.Context, userID string, question string) ([]fact.FactListView, error) {
	embeddings, err := r.e.GetEmbeddings(ctx, question)
	if err != nil {
		r.logger.Error("get question embeddings failed", "user_id", userID, "question_length", len(question), "error", err)

		return []fact.FactListView{}, err
	}

	if len(embeddings) == 0 {
		r.logger.Error("question embeddings are empty", "user_id", userID, "question_length", len(question), "error", apperrors.ErrEmptyEmbeddings)

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
		r.logger.Error("get facts list failed", "user_id", userID, "top_k", r.c.GetTopK(), "threshold", r.c.GetSimilarityThreshold(), "error", err)

		return []fact.FactListView{}, err
	}

	return facts, nil
}
