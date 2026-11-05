package kv

import (
	"context"
	"fmt"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

func (k *KV) GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error) {
	if !validFactID(factID) {
		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	collection, err := k.getCollection(ctx)
	if err != nil {
		return fact.FactFullView{}, err
	}
	defer collection.Close()

	result, err := collection.Fetch(
		ctx,
		[]string{factID},
		xvec.Projection{
			OutputFields: []string{keyFieldCreatedAt, keyFieldSummary, keyFieldFact},
		},
	)
	if err != nil {
		return fact.FactFullView{}, err
	}

	if len(result) == 0 || result[0] == nil {
		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	doc := result[0]

	createdAt, ok := doc.Fields[keyFieldCreatedAt].(string)
	if !ok {
		return fact.FactFullView{}, fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldCreatedAt,
			doc.Fields[keyFieldCreatedAt],
		)
	}

	summary, ok := doc.Fields[keyFieldSummary].(string)
	if !ok {
		return fact.FactFullView{}, fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldSummary,
			doc.Fields[keyFieldSummary],
		)
	}

	content, ok := doc.Fields[keyFieldFact].(string)
	if !ok {
		return fact.FactFullView{}, fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldFact,
			doc.Fields[keyFieldFact],
		)
	}

	return fact.FactFullView{
		ID:        factID,
		CreatedAt: createdAt,
		Summary:   summary,
		Content:   content,
	}, nil
}
