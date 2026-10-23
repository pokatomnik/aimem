package removefact

import "context"

type (
	store interface {
		RemoveFactByID(ctx context.Context, id string) error
	}

	RemoveFact struct {
		s store
	}
)

func New(s store) *RemoveFact {
	return &RemoveFact{s}
}

func (r *RemoveFact) RemoveFactByID(ctx context.Context, id string) error {
	return r.s.RemoveFactByID(ctx, id)
}
