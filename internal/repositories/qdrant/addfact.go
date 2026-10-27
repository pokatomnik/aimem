package qdrant

import (
	"context"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// AddFact stores one fact with its vector and summary, and reports a fact that is
// already stored as apperrors.ErrAlreadyHasFact.
//
// Qdrant's upsert endpoint overwrites an existing point, while kv rejects the
// duplicate; the existence check is what reproduces kv's "already has fact"
// answer. The check and the upsert are two separate requests, so two concurrent
// writers of the same text can both pass the check and both upsert. That is
// harmless: the point ID is a pure function of the text, so the loser rewrites
// the same fact, and the later write only refreshes summary and created_at —
// content and identity never change for a given ID. The write is what makes the
// race benign, so no lock is added here; a process-local lock would not close the
// window for a second instance anyway.
func (q *Qdrant) AddFact(ctx context.Context, vector []float32, summary, content string) error {
	if err := q.ensureCollection(ctx); err != nil {
		return err
	}

	id := newPointID(content)

	exists, err := q.pointExists(ctx, id)
	if err != nil {
		return err
	}

	if exists {
		return apperrors.ErrAlreadyHasFact
	}

	request := upsertPointsRequest{
		Points: []pointStruct{{
			ID:     id,
			Vector: namedVector(vector),
			Payload: map[string]any{
				keyFieldCreatedAt: now(),
				keyFieldSummary:   summary,
				keyFieldFact:      content,
			},
		}},
	}

	return q.do(ctx, http.MethodPut, pathPoints, writeQuery(), request, nil)
}

// pointExists reports whether a point with this ID is stored.
func (q *Qdrant) pointExists(ctx context.Context, id string) (bool, error) {
	request := retrievePointsRequest{
		IDs:         []string{id},
		WithPayload: false,
		WithVector:  false,
	}

	var points []retrievedPoint

	if err := q.do(ctx, http.MethodPost, pathPoints, nil, request, &points); err != nil {
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
