package kv

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func (k *KV) RemoveFactByID(ctx context.Context, id string) error {
	result, err := k.collection.Delete(ctx, []string{id})
	if err != nil {
		return err
	}

	for _, v := range result {
		if v.Err != nil {
			return err
		}
	}

	if len(result) == 0 {
		return apperrors.ErrFactDoesNotExist
	}

	return nil
}
