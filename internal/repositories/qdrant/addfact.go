package qdrant

import (
	"context"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// AddFact stores one fact with its vector and summary, and reports a fact that is
// already stored as apperrors.ErrAlreadyHasFact.
//
// The fact is written to the caller's own collection, so "already stored" means
// stored by this user: the same text from another user is a different record and
// is accepted. What the duplicate check compares is therefore scoped to the user,
// not the server.
//
// Qdrant's upsert endpoint overwrites an existing point, while kv rejects the
// duplicate; the existence check is what reproduces kv's "already has fact"
// answer. The check and the upsert are two separate requests, so two concurrent
// writers of the same text can both pass the check and both upsert. That is
// harmless: the point ID is a pure function of the text, so the loser rewrites
// the same fact, and the later write only refreshes summary and created_at —
// content and identity never change for a given ID. The write is what makes the
// race benign, so no lock is added here; a process-local lock would not close the
// window for a second instance anyway, and it would have to be per collection to
// mean anything now that one server holds several users.
func (q *Qdrant) AddFact(ctx context.Context, userID string, vector []float32, summary, content string) error {
	if err := q.ensureCollection(ctx, userID); err != nil {
		return err
	}

	factID := newPointID(content)

	exists, err := q.pointExists(ctx, userID, factID)
	if err != nil {
		return err
	}

	if exists {
		return apperrors.ErrAlreadyHasFact
	}

	request := upsertPointsRequest{
		Points: []pointStruct{{
			ID:     factID,
			Vector: namedVector(vector),
			Payload: map[string]any{
				keyFieldCreatedAt: now(),
				keyFieldSummary:   summary,
				keyFieldFact:      content,
			},
		}},
	}

	return q.do(ctx, http.MethodPut, pathPoints(userID), writeQuery(), request, nil)
}

// pointExists reports whether a point with this ID is stored in the user's
// collection. A fact another user holds is not visible here, which is what makes
// a foreign ID look absent rather than forbidden.
func (q *Qdrant) pointExists(ctx context.Context, userID, factID string) (bool, error) {
	request := retrievePointsRequest{
		IDs:         []string{factID},
		WithPayload: false,
		WithVector:  false,
	}

	var points []retrievedPoint

	if err := q.do(ctx, http.MethodPost, pathPoints(userID), nil, request, &points); err != nil {
		return false, err
	}

	return len(points) > 0, nil
}

// namedVector wraps a dense vector under the collection's vector name. The
// collection declares a named vector, so a bare array would address Qdrant's
// default vector "", which does not exist there.
func namedVector(vector []float32) map[string][]float32 {
	return map[string][]float32{keyFieldEmbedding: vector}
}
