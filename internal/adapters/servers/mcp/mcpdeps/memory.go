package mcpdeps

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	// Authenticator turns the credentials of a tool call into the user it belongs
	// to. It is the only thing a tool knows about authentication: whether those
	// credentials are required at all was decided when the server was built, and is
	// expressed by which Authenticator was injected.
	//
	// An anonymous implementation returns an empty ID without reading anything, so a
	// tool holds no branch of its own and cannot disagree with the web API about the
	// mode this deployment runs in.
	Authenticator interface {
		UserID(ctx context.Context, req *mcp.CallToolRequest) (string, error)
	}
)
