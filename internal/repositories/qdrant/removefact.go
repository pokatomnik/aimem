package qdrant

import (
	"context"
	"net/http"
)

// RemoveFactByID deletes one fact. Deleting a fact that is not stored succeeds:
// DELETE is idempotent, so a client can retry it and gets the same answer as the
// first call. Qdrant's delete endpoint already answers that way, which is why the
// existence check this method used to make is gone — it existed only to turn a
// missing fact into an error.
//
// Only the caller's collection is touched, so an ID held by another user deletes
// nothing and is reported as success, exactly like an ID nobody stored.
//
// An ID that cannot be mapped to a point is a no-op as well, and is answered
// before the collection is provisioned: no collection could hold such a fact.
func (q *Qdrant) RemoveFactByID(ctx context.Context, userID, factID string) error {
	pointID, err := pointIDFromExternal(factID)
	if err != nil {
		return nil
	}

	if err := q.ensureCollection(ctx, userID); err != nil {
		q.logger.Error("ensure collection failed", "user_id", userID, "error", err)

		return err
	}

	request := deletePointsRequest{Points: []string{pointID}}

	if err := q.do(ctx, http.MethodPost, pathPointsDelete(userID), writeQuery(), request, nil); err != nil {
		q.logger.Error("delete fact failed", "user_id", userID, "fact_id", factID, "error", err)

		return err
	}

	return nil
}
