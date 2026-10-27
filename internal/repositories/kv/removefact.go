package kv

import (
	"context"
	"errors"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func (k *KV) RemoveFactByID(ctx context.Context, id string) error {
	collection, err := k.getCollection(ctx)
	if err != nil {
		return err
	}
	defer collection.Close()

	result, err := collection.Delete(ctx, []string{id})
	// xvec reports a missing key as a batch error carrying ErrNotFound, and puts
	// the same cause into the per-key result. Both mean "no such fact".
	if errors.Is(err, xvec.ErrNotFound) {
		return apperrors.ErrFactDoesNotExist
	}

	if err != nil {
		return err
	}

	for _, v := range result {
		if errors.Is(v.Err, xvec.ErrNotFound) {
			return apperrors.ErrFactDoesNotExist
		}

		if v.Err != nil {
			return v.Err
		}
	}

	if len(result) == 0 {
		return apperrors.ErrFactDoesNotExist
	}

	return nil
}
