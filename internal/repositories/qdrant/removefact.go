package qdrant

import (
	"context"
	"net/http"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// RemoveFactByID deletes one fact and reports an unknown ID as
// apperrors.ErrFactDoesNotExist, the same way kv does.
//
// Only the caller's collection is touched, so an ID held by another user is not
// found here and is reported as missing instead of being deleted.
//
// The delete endpoint succeeds whether or not the point was there, so the check
// runs first — the same order kv uses, where Delete returns an error per missing
// key.
func (q *Qdrant) RemoveFactByID(ctx context.Context, userID, factID string) error {
	if err := q.ensureCollection(ctx, userID); err != nil {
		return err
	}

	pointID, err := pointIDFromExternal(factID)
	if err != nil {
		return err
	}

	exists, err := q.pointExists(ctx, userID, pointID)
	if err != nil {
		return err
	}

	if !exists {
		return apperrors.ErrFactDoesNotExist
	}

	request := deletePointsRequest{Points: []string{pointID}}

	return q.do(ctx, http.MethodPost, pathPointsDelete(userID), writeQuery(), request, nil)
}
