package qdrant

import (
	"context"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// RemoveFactByID deletes one fact and reports an unknown ID as
// apperrors.ErrFactDoesNotExist, the same way kv does.
//
// The delete endpoint succeeds whether or not the point was there, so the check
// runs first — the same order kv uses, where Delete returns an error per missing
// key.
func (q *Qdrant) RemoveFactByID(ctx context.Context, id string) error {
	if err := q.ensureCollection(ctx); err != nil {
		return err
	}

	pointID, err := pointIDFromExternal(id)
	if err != nil {
		return err
	}

	exists, err := q.pointExists(ctx, pointID)
	if err != nil {
		return err
	}

	if !exists {
		return apperrors.ErrFactDoesNotExist
	}

	request := deletePointsRequest{Points: []string{pointID}}

	return q.do(ctx, http.MethodPost, pathPointsDelete, writeQuery(), request, nil)
}
