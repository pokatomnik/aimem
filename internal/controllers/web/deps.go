package web

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
)

type (
	Config interface {
		GetWebHTTPHost() string
	}

	GetFact interface {
		GetFact(ctx context.Context, id string) (fact.FactFullView, error)
	}

	Recall interface {
		Recall(ctx context.Context, question string) ([]fact.FactListView, error)
	}

	Remember interface {
		Remember(ctx context.Context, fact string) (string, error)
	}

	RemoveFact interface {
		RemoveFactByID(ctx context.Context, id string) error
	}
)
