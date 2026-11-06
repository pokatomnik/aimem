package removefact

import (
	"context"

	"github.com/pokatomnik/gomem/internal/shared/logging"
)

type (
	store interface {
		RemoveFactByID(ctx context.Context, userID, factID string) error
	}

	RemoveFact struct {
		s      store
		logger logging.Logger
	}
)

func New(s store, logger logging.Logger) *RemoveFact {
	return &RemoveFact{s: s, logger: logger}
}

func (r *RemoveFact) RemoveFactByID(ctx context.Context, userID, factID string) error {
	err := r.s.RemoveFactByID(ctx, userID, factID)
	if err != nil {
		r.logger.Error("remove fact failed", "user_id", userID, "fact_id", factID, "error", err)
	}

	return err
}
