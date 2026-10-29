package mcpdeps

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

type (
	Remember interface {
		Remember(ctx context.Context, userID, fact string) (string, error)
	}

	Recall interface {
		Recall(ctx context.Context, userID, question string) ([]fact.FactListView, error)
	}

	GetFact interface {
		GetFact(ctx context.Context, userID, factID string) (fact.FactFullView, error)
	}

	RemoveFact interface {
		RemoveFactByID(ctx context.Context, userID, factID string) error
	}

	Authentication interface {
		GetUserByToken(ctx context.Context, token string) (user.User, error)
	}
)
