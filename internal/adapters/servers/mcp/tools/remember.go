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
		Description: "Save a self-contained user-provided fact; return a short summary. " +
			"Then use recall and read_fact to check conflicts.",
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
