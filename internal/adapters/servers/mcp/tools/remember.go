package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	rememberInput struct {
		Fact string `json:"fact" jsonschema:"Fact/knowledge to remember"`
	}
	rememberOutput struct {
		Summary string `json:"summary" jsonschema:"Remembered fact/knowledge summary"`
	}
)

func ToolRemember(a mcpdeps.Authenticator, r mcpdeps.Remember, l mcpdeps.Logger) (
	mcp.Tool,
	mcp.ToolHandlerFor[rememberInput, rememberOutput],
	error,
) {
	tool := mcp.Tool{
		Name: "remember",
		Description: "Save a fact or piece of knowledge to persistent memory. The result is the short summary that " +
			"the recall tool will later show, so write the fact as a complete self-contained statement. Save only what " +
			"the user actually provided; this is not a transcript.",
	}

	var handler mcp.ToolHandlerFor[rememberInput, rememberOutput] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input rememberInput,
	) (
		*mcp.CallToolResult,
		rememberOutput,
		error,
	) {
		userID, err := a.UserID(ctx, req)
		if err != nil {
			logToolError(l, tool.Name, "", err)

			return &mcp.CallToolResult{IsError: true}, rememberOutput{}, err
		}

		summary, err := r.Remember(ctx, userID, input.Fact)
		if err != nil {
			logToolError(l, tool.Name, userID, err)

			return &mcp.CallToolResult{IsError: true}, rememberOutput{}, err
		}

		return nil, rememberOutput{
			Summary: summary,
		}, nil
	}

	return tool, handler, nil
}
