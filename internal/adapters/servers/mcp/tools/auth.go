package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/shared/server"
)

// authenticate resolves the caller from the Authorization header of the tool
// call and returns the user's ID.
//
// The header arrives through the transport, not through the tool arguments: the
// streamable HTTP transport fills RequestExtra.Header, so a call made over any
// other transport (the in-memory one in tests, for instance) carries no Extra at
// all. That case must be a plain "no credentials" answer — dereferencing it
// would panic inside the handler and leave the call unanswered instead of
// returning an error.
func authenticate(
	ctx context.Context,
	req *mcp.CallToolRequest,
	a mcpdeps.Authentication,
) (string, error) {
	if req == nil || req.Extra == nil {
		return "", apperrors.ErrUserNotFound
	}

	authHeaderVal := req.Extra.Header.Get(server.AuthorizationHeader)
	if authHeaderVal == "" {
		return "", apperrors.ErrUserNotFound
	}

	token, err := server.ParseAuthHeader(authHeaderVal)
	if err != nil {
		return "", apperrors.ErrUserNotFound
	}

	u, err := a.GetUserByToken(ctx, token)
	if err != nil {
		return "", apperrors.ErrUserNotFound
	}

	return u.ID, nil
}
