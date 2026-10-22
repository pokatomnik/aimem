package mcpdeps

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
)

type (
	Remember interface {
		Remember(ctx context.Context, fact string) (string, error)
	}

	Recall interface {
		Recall(ctx context.Context, question string) ([]fact.FactListView, error)
	}

	GetFact interface {
		GetFact(ctx context.Context, id string) (fact.FactFullView, error)
	}
)
