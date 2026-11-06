package tools

import (
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// logToolError records a failed tool call once, at the edge.
//
// The use cases stay logger-free and just return errors, so this is where a
// failure gets the context only the tool knows — its name. A refusal caused by
// the caller or its data is a warning; a fault on our side is an error. A nil
// logger is tolerated because a panic here would leave the call unanswered
// instead of telling the caller it failed.
func logToolError(l mcpdeps.Logger, tool, userID string, err error) {
	if l == nil {
		return
	}

	args := []any{"tool", tool, "user_id", userID, "error", err}

	if apperrors.IsClientError(err) {
		l.Warn("tool call rejected", args...)

		return
	}

	l.Error("tool call failed", args...)
}
