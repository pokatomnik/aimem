package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/shared/server"
)

// NewTokenAuthenticator returns the authenticator used when every call must carry
// a valid token.
func NewTokenAuthenticator(users mcpdeps.Authentication, logger mcpdeps.Logger) mcpdeps.Authenticator {
	return tokenAuthenticator{users: users, logger: logger}
}

// NewAnonymousAuthenticator returns the authenticator for the embedded backend,
// which ignores the user ID and therefore requires no credentials.
func NewAnonymousAuthenticator() mcpdeps.Authenticator {
	return anonymousAuthenticator{}
}

// tokenAuthenticator resolves the bearer token of a call into a user ID.
//
// It reads the header from the transport, not from the tool arguments: the
// streamable HTTP transport fills RequestExtra.Header, so a call made over any
// other transport (the in-memory one in tests, for instance) carries no Extra at
// all. That case must be a plain "no credentials" answer — dereferencing it would
// panic inside the handler and leave the call unanswered instead of returning an
// error.
type tokenAuthenticator struct {
	users  mcpdeps.Authentication
	logger mcpdeps.Logger
}

func (a tokenAuthenticator) UserID(ctx context.Context, req *mcp.CallToolRequest) (string, error) {
	tool := ""
	if req != nil && req.Params != nil {
		tool = req.Params.Name
	}

	if req == nil || req.Extra == nil {
		a.logger.Warn("tool call has no request headers", "tool", tool, "error", apperrors.ErrUserNotFound)

		return "", apperrors.ErrUserNotFound
	}

	authHeaderVal := req.Extra.Header.Get(server.AuthorizationHeader)
	if authHeaderVal == "" {
		a.logger.Warn("tool call has no bearer token", "tool", tool, "error", apperrors.ErrUserNotFound)

		return "", apperrors.ErrUserNotFound
	}

	token, err := server.ParseAuthHeader(authHeaderVal)
	if err != nil {
		a.logger.Warn("tool call has a malformed bearer token", "tool", tool, "error", err)

		return "", apperrors.ErrUserNotFound
	}

	u, err := a.users.GetUserByToken(ctx, token)
	if err != nil {
		a.logger.Warn("tool call has an unknown token", "tool", tool, "error", err)

		return "", apperrors.ErrUserNotFound
	}

	return u.ID, nil
}

// anonymousAuthenticator serves the embedded backend, which stores one operator's
// local files and ignores the user ID. It reads nothing and refuses nothing, so a
// tool needs no branch to tell the two modes apart.
type anonymousAuthenticator struct{}

func (anonymousAuthenticator) UserID(context.Context, *mcp.CallToolRequest) (string, error) {
	return "", nil
}
