package getfact

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
)

type (
	store interface {
		GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error)
	}

	GetFact struct {
		s store
	}
)

func New(s store) *GetFact {
	return &GetFact{s}
}

func (r *GetFact) GetFact(ctx context.Context, userID, factID string) (fact.FactFullView, error) {
	return r.s.GetFactByID(ctx, userID, factID)
}
