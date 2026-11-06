package kv

import (
	"context"
	"errors"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func (k *KV) AddFact(ctx context.Context, userID string, vector []float32, summary, content string) error {
	id := k.hash(content)

	document, err := xvec.NewDocument(id, map[string]any{
		keyFieldCreatedAt: k.now(),
		keyFieldSummary:   summary,
		keyFieldFact:      content,
		keyFieldEmbedding: xvec.VectorFP32(vector),
	})
	if err != nil {
		k.logger.Error("build fact document failed", "user_id", userID, "content_length", len(content), "error", err)

		return err
	}

	collection, err := k.getCollection(ctx)
	if err != nil {
		k.logger.Error("open collection failed", "user_id", userID, "path", k.dataPath, "error", err)

		return err
	}
	defer collection.Close()

	results, err := collection.Insert(ctx, []xvec.Document{document})
	if errors.Is(err, xvec.ErrAlreadyExists) {
		k.logger.Warn("fact already exists", "user_id", userID, "fact_id", id, "error", err)

		return apperrors.ErrAlreadyHasFact
	}

	if err != nil {
		k.logger.Error("insert fact failed", "user_id", userID, "fact_id", id, "error", err)

		return err
	}

	for _, result := range results {
		if result.Err != nil {
			k.logger.Error("insert fact failed", "user_id", userID, "fact_id", id, "error", result.Err)

			return result.Err
		}
	}

	return nil
}
