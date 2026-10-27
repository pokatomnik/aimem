package qdrant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

// GetFactByID returns one stored fact in full.
func (q *Qdrant) GetFactByID(ctx context.Context, id string) (fact.FactFullView, error) {
	if err := q.ensureCollection(ctx); err != nil {
		return fact.FactFullView{}, err
	}

	pointID, err := pointIDFromExternal(id)
	if err != nil {
		return fact.FactFullView{}, err
	}

	request := retrievePointsRequest{
		IDs:         []string{pointID},
		WithPayload: []string{keyFieldCreatedAt, keyFieldSummary, keyFieldFact},
		WithVector:  false,
	}

	var points []retrievedPoint

	if err := q.do(ctx, http.MethodPost, pathPoints, nil, request, &points); err != nil {
		return fact.FactFullView{}, err
	}

	if len(points) == 0 {
		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	payload := points[0].Payload

	createdAt, err := stringField(payload, keyFieldCreatedAt)
	if err != nil {
		return fact.FactFullView{}, err
	}

	summary, err := stringField(payload, keyFieldSummary)
	if err != nil {
		return fact.FactFullView{}, err
	}

	content, err := stringField(payload, keyFieldFact)
	if err != nil {
		return fact.FactFullView{}, err
	}

	return fact.FactFullView{
		ID:        points[0].ID.UUID,
		CreatedAt: createdAt,
		Summary:   summary,
		Content:   content,
	}, nil
}

// stringField reads a payload field and keeps kv's conversion contract: a field
// that is missing or not a string is ErrDBConvertError, not a zero value.
func stringField(payload map[string]any, key string) (string, error) {
	value, ok := payload[key].(string)
	if !ok {
		return "", fmt.Errorf(
			"%w: key: %s, value: %v is not of type string",
			apperrors.ErrDBConvertError,
			key,
			payload[key],
		)
	}

	return value, nil
}
