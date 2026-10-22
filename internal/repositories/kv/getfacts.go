package kv

import (
	"context"
	"fmt"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

func (k *KV) GetFactsList(ctx context.Context, vector []float32, topK int) (facts []fact.FactListView, err error) {
	results, err := k.collection.Query(ctx, xvec.VectorQuery{
		Field:       keyFieldEmbedding,
		DenseVector: xvec.VectorFP32(vector),
		TopK:        topK,
		Projection: xvec.Projection{
			OutputFields: []string{keyFieldCreatedAt, keyFieldSummary},
		},
	})
	if err != nil {
		return nil, err
	}

	facts = make([]fact.FactListView, 0, len(results))
	for _, result := range results {
		createdAt, ok := result.Fields[keyFieldCreatedAt].(string)
		if !ok {
			return nil, fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrorDBConvertError,
				keyFieldCreatedAt,
				result.Fields[keyFieldCreatedAt],
			)
		}

		summary, ok := result.Fields[keyFieldSummary].(string)
		if !ok {
			return nil, fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrorDBConvertError,
				keyFieldSummary,
				result.Fields[keyFieldSummary],
			)
		}

		facts = append(facts, fact.FactListView{
			ID:        result.PrimaryKey,
			CreatedAt: createdAt,
			Summary:   summary,
		})
	}

	return facts, nil
}
