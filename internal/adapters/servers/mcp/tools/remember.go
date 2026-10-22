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

func ToolRemember(r mcpdeps.Remember) (
	mcp.Tool,
	mcp.ToolHandlerFor[rememberInput, rememberOutput],
	error,
) {
	tool := mcp.Tool{
		Name:        "remember",
		Description: "Remember a fact/knowledge",
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
		summary, err := r.Remember(ctx, input.Fact)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, rememberOutput{}, err
		}

		return nil, rememberOutput{
			Summary: summary,
		}, nil
	}

	return tool, handler, nil
}
