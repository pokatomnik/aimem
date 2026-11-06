package kv

import (
	"context"
	"fmt"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

func (k *KV) GetFactsList(ctx context.Context, userID string, vector []float32, topK int, similarityThreshold float32) (facts []fact.FactListView, err error) {
	collection, err := k.getCollection(ctx)
	if err != nil {
		k.logger.Error("open collection failed", "user_id", userID, "path", k.dataPath, "error", err)

		return []fact.FactListView{}, err
	}
	defer collection.Close()

	results, err := collection.Query(ctx, xvec.VectorQuery{
		Field:       keyFieldEmbedding,
		DenseVector: xvec.VectorFP32(vector),
		TopK:        topK,
		Projection: xvec.Projection{
			OutputFields: []string{keyFieldCreatedAt, keyFieldSummary},
		},
	})
	if err != nil {
		k.logger.Error("query facts failed", "user_id", userID, "top_k", topK, "threshold", similarityThreshold, "error", err)

		return nil, err
	}

	facts = make([]fact.FactListView, 0, len(results))
	for _, result := range results {
		// xvec reports the cosine metric as a distance: 0 means identical, 1 means
		// orthogonal. The config threshold is a similarity, so flip the direction
		// before comparing, otherwise the closest facts get filtered out.
		similarity := 1 - result.Score
		if similarity < similarityThreshold {
			continue
		}

		createdAt, ok := result.Fields[keyFieldCreatedAt].(string)
		if !ok {
			err := fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrDBConvertError,
				keyFieldCreatedAt,
				result.Fields[keyFieldCreatedAt],
			)
			k.logger.Error("fact created_at is not a string", "user_id", userID, "fact_id", result.PrimaryKey, "error", err)

			return nil, err
		}

		summary, ok := result.Fields[keyFieldSummary].(string)
		if !ok {
			err := fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrDBConvertError,
				keyFieldSummary,
				result.Fields[keyFieldSummary],
			)
			k.logger.Error("fact summary is not a string", "user_id", userID, "fact_id", result.PrimaryKey, "error", err)

			return nil, err
		}

		facts = append(facts, fact.FactListView{
			ID:        result.PrimaryKey,
			CreatedAt: createdAt,
			Summary:   summary,
		})
	}

	return facts, nil
}
