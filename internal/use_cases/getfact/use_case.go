package getfact

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/shared/logging"
)

type (
	store interface {
		GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error)
	}

	GetFact struct {
		s      store
		logger logging.Logger
	}
)

func New(s store, logger logging.Logger) *GetFact {
	return &GetFact{s: s, logger: logger}
}

func (r *GetFact) GetFact(ctx context.Context, userID, factID string) (fact.FactFullView, error) {
	view, err := r.s.GetFactByID(ctx, userID, factID)
	if err != nil {
		r.logger.Error("get fact failed", "user_id", userID, "fact_id", factID, "error", err)
	}

	return view, err
}
