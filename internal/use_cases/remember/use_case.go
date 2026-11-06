package remember

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/embeddings"
	"github.com/pokatomnik/gomem/internal/shared/logging"
)

type (
	embedder interface {
		GetEmbeddings(ctx context.Context, text string) (embeddings.Embeddings, error)
	}

	summarizer interface {
		Summarize(ctx context.Context, text string) (string, error)
	}

	store interface {
		AddFact(ctx context.Context, userID string, vector []float32, summary, content string) error
	}

	Remember struct {
		embedder   embedder
		summarizer summarizer
		st         store
		logger     logging.Logger
	}
)

func New(embedder embedder, summarizer summarizer, s store, logger logging.Logger) *Remember {
	return &Remember{
		embedder:   embedder,
		summarizer: summarizer,
		st:         s,
		logger:     logger,
	}
}

func (r *Remember) Remember(ctx context.Context, userID string, fact string) (string, error) {
	summary, embeddings, err := r.getEmbeddingsAndSummary(ctx, fact)
	if err != nil {
		r.logger.Error("summarize and embed fact failed", "user_id", userID, "fact_length", len(fact), "error", err)

		return "", err
	}

	if len(embeddings) != 1 {
		r.logger.Error("embeddings quantity does not equal facts quantity", "user_id", userID, "fact_length", len(fact), "embeddings", len(embeddings), "error", apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity)

		return "", apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity
	}

	embedding := embeddings[0]
	err = r.st.AddFact(ctx, userID, embedding, summary, fact)

	if err != nil {
		r.logger.Error("add fact failed", "user_id", userID, "fact_length", len(fact), "error", err)

		return "", err
	}

	return summary, nil
}
