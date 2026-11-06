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
		k.logger.Warn("fact id cannot name a stored fact", "user_id", userID, "fact_id", factID, "id_length", len(factID), "error", apperrors.ErrFactDoesNotExist)

		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	collection, err := k.getCollection(ctx)
	if err != nil {
		k.logger.Error("open collection failed", "user_id", userID, "path", k.dataPath, "error", err)

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
		k.logger.Error("fetch fact failed", "user_id", userID, "fact_id", factID, "error", err)

		return fact.FactFullView{}, err
	}

	if len(result) == 0 || result[0] == nil {
		k.logger.Warn("fact does not exist", "user_id", userID, "fact_id", factID, "error", apperrors.ErrFactDoesNotExist)

		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	doc := result[0]

	createdAt, ok := doc.Fields[keyFieldCreatedAt].(string)
	if !ok {
		err := fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldCreatedAt,
			doc.Fields[keyFieldCreatedAt],
		)
		k.logger.Error("fact created_at is not a string", "user_id", userID, "fact_id", factID, "error", err)

		return fact.FactFullView{}, err
	}

	summary, ok := doc.Fields[keyFieldSummary].(string)
	if !ok {
		err := fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldSummary,
			doc.Fields[keyFieldSummary],
		)
		k.logger.Error("fact summary is not a string", "user_id", userID, "fact_id", factID, "error", err)

		return fact.FactFullView{}, err
	}

	content, ok := doc.Fields[keyFieldFact].(string)
	if !ok {
		err := fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			keyFieldFact,
			doc.Fields[keyFieldFact],
		)
		k.logger.Error("fact content is not a string", "user_id", userID, "fact_id", factID, "error", err)

		return fact.FactFullView{}, err
	}

	return fact.FactFullView{
		ID:        factID,
		CreatedAt: createdAt,
		Summary:   summary,
		Content:   content,
	}, nil
}
