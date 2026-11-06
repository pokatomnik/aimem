package web

import (
	"context"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/user"
	"github.com/pokatomnik/gomem/internal/shared/logging"
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

	// logger keeps the local spelling used by this package while the interface
	// itself lives in shared, where every layer can depend on it.
	logger = logging.Logger
)
