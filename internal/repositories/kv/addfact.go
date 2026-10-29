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
		return err
	}

	collection, err := k.getCollection(ctx)
	if err != nil {
		return err
	}
	defer collection.Close()

	results, err := collection.Insert(ctx, []xvec.Document{document})
	if errors.Is(err, xvec.ErrAlreadyExists) {
		return apperrors.ErrAlreadyHasFact
	}

	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Err != nil {
			return result.Err
		}
	}

	return k.appendPlainTextMessage(content)
}
