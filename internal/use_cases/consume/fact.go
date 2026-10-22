package consume

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func (d *Consumer) addFact(ctx context.Context, fact string) error {
	summary, embeddings, err := d.getEmbeddingsAndSummary(ctx, fact)
	if err != nil {
		return err
	}

	if len(embeddings) != 1 {
		return apperrors.ErrEmbeddingsQuantityNotEqualToFactsQuantity
	}

	embedding := embeddings[0]
	err = d.s.AddFact(ctx, embedding, summary, fact)

	if err != nil {
		return err
	}

	return nil
}
