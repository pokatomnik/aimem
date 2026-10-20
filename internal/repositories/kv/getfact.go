package kv

import (
	"context"
	"fmt"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

func (k *KV) GetFactByID(ctx context.Context, id string) (fact.FactFullView, error) {
	result, err := k.collection.Fetch(
		ctx,
		[]string{id},
		xvec.Projection{
			OutputFields: []string{keyFieldSummary, keyFieldFact},
		},
	)
	if err != nil {
		return fact.FactFullView{}, err
	}

	if len(result) == 0 {
		return fact.FactFullView{}, apperrors.ErrorFactDoesNotExist
	}

	doc := result[0]

	summary, ok := doc.Fields[keyFieldSummary].(string)
	if !ok {
		return fact.FactFullView{}, fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrorDBConvertError,
			keyFieldSummary,
			doc.Fields[keyFieldSummary],
		)
	}

	content, ok := doc.Fields[keyFieldFact].(string)
	if !ok {
		return fact.FactFullView{}, fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrorDBConvertError,
			keyFieldFact,
			doc.Fields[keyFieldFact],
		)
	}

	return fact.FactFullView{
		ID:      id,
		Summary: summary,
		Content: content,
	}, nil
}
