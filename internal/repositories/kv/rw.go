package kv

import (
	"context"
	"fmt"
	"uuid"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

func (k *KV) AddFact(ctx context.Context, vector []float32, content string) error {
	id := uuid.New()

	document, err := xvec.NewDocument(id.String(), map[string]any{
		"id":              id.String(),
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
			return err
		}
	}

	return nil
}

func (k *KV) GetFacts(ctx context.Context, vector []float32, topK int) (facts []fact.Fact, err error) {
	results, err := k.collection.Query(ctx, xvec.VectorQuery{
		Field:       keyFieldEmbedding,
		DenseVector: xvec.VectorFP32(vector),
		TopK:        topK,
		Projection: xvec.Projection{
			OutputFields: []string{keyFieldID, keyFieldFact},
		},
	})
	if err != nil {
		return nil, err
	}

	facts = make([]fact.Fact, 0, len(results))
	for _, result := range results {
		id, ok := result.Fields[keyFieldID].(string)
		if !ok {
			return nil, fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrorDBConvertError,
				keyFieldID,
				result.Fields[keyFieldID],
			)
		}

		content, ok := result.Fields[keyFieldFact].(string)
		if !ok {
			return nil, fmt.Errorf(
				"%w: key: %s, value: %v is not of type string",
				apperrors.ErrorDBConvertError,
				keyFieldFact,
				result.Fields[keyFieldFact],
			)
		}

		facts = append(facts, fact.Fact{
			ID:      id,
			Content: content,
		})
	}

	return facts, nil
}
