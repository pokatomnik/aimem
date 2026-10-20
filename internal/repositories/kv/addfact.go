package kv

import (
	"context"
	"uuid"

	"github.com/gorse-io/xvec"
)

func (k *KV) AddFact(ctx context.Context, vector []float32, summary, content string) error {
	id := uuid.New()

	document, err := xvec.NewDocument(id.String(), map[string]any{
		keyFieldSummary:   summary,
		keyFieldFact:      content,
		keyFieldEmbedding: xvec.VectorFP32(vector),
	})
	if err != nil {
		return err
	}

	results, err := k.collection.Insert(ctx, []xvec.Document{document})
	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Err != nil {
			return result.Err
		}
	}

	return nil
}
