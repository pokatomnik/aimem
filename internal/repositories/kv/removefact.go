package kv

import (
	"context"
	"errors"

	"github.com/gorse-io/xvec"
)

// RemoveFactByID deletes one fact. Deleting a fact that is not stored succeeds:
// DELETE is idempotent, so a client can retry it and gets the same answer as the
// first call. Nothing in xvec distinguishes "removed just now" from "was not
// there", and the caller has no use for the distinction.
//
// An ID the index cannot hold is a no-op too, and is answered before the
// collection is opened: xvec rejects an empty, non-UTF-8 or oversized primary key
// with a validation error, which would turn a guaranteed no-op into a 500.
func (k *KV) RemoveFactByID(ctx context.Context, userID, factID string) error {
	if !validFactID(factID) {
		return nil
	}

	collection, err := k.getCollection(ctx)
	if err != nil {
		k.logger.Error("open collection failed", "user_id", userID, "path", k.dataPath, "error", err)

		return err
	}
	defer collection.Close()

	results, err := collection.Delete(ctx, []string{factID})
	// xvec reports a missing key as a batch error carrying ErrNotFound, and puts
	// the same cause into the per-key result. Both mean "nothing to delete".
	if err != nil && !errors.Is(err, xvec.ErrNotFound) {
		k.logger.Error("delete fact failed", "user_id", userID, "fact_id", factID, "error", err)

		return err
	}

	for _, result := range results {
		if result.Err != nil && !errors.Is(result.Err, xvec.ErrNotFound) {
			k.logger.Error("delete fact failed", "user_id", userID, "fact_id", factID, "error", result.Err)

			return result.Err
		}
	}

	return nil
}
