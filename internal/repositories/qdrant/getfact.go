package qdrant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

// GetFactByID returns one stored fact in full.
//
// The lookup runs against the caller's own collection, so an ID that belongs to
// another user is not found and comes back as apperrors.ErrFactDoesNotExist. That
// is deliberate: a different user's fact should be indistinguishable from one
// that was never stored, so the answer leaks nothing about what exists elsewhere.
func (q *Qdrant) GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error) {
	// Qdrant accepts a point ID only as a UUID or a u64, so a string this package
	// cannot map cannot name a point in any collection. That is a fact that does
	// not exist, not a conversion failure: the caller may hold an ID minted while
	// the embedded backend was selected, or a hand-edited URL. Answering "absent"
	// keeps this indistinguishable from a foreign user's ID, which is how kv
	// answers the same input. The check runs before the bootstrap because no
	// collection is worth creating for an ID that cannot be stored.
	pointID, err := pointIDFromExternal(factID)
	if err != nil {
		return fact.FactFullView{}, apperrors.ErrFactDoesNotExist
	}

	if err := q.ensureCollection(ctx, userID); err != nil {
		return fact.FactFullView{}, err
	}

	request := retrievePointsRequest{
		IDs:         []string{pointID},
		WithPayload: []string{keyFieldCreatedAt, keyFieldSummary, keyFieldFact},
		WithVector:  false,
	}

	var points []retrievedPoint

	if err := q.do(ctx, http.MethodPost, pathPoints(userID), nil, request, &points); err != nil {
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
