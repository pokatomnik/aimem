package web

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

type (
	Config interface {
		GetWebHTTPHost() string
	}

	GetFact interface {
		GetFact(ctx context.Context, userID, factID string) (fact.FactFullView, error)
	}

	Recall interface {
		Recall(ctx context.Context, userID, question string) ([]fact.FactListView, error)
	}

	Remember interface {
		Remember(ctx context.Context, userID, fact string) (string, error)
	}

	RemoveFact interface {
		RemoveFactByID(ctx context.Context, userID, factID string) error
	}

	Authentication interface {
		GetUserByToken(ctx context.Context, token string) (user.User, error)
	}
)
