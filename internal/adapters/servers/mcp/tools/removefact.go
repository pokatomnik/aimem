package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	removeFactInput struct {
		ID string `json:"id" jsonschema:"Fact/knowledge ID"`
	}

	removeFactOutput struct {
		Ok bool `json:"ok" jsonschema:"Ok is true if fact removed"`
	}
)

func ToolRemoveFact(r mcpdeps.RemoveFact) (
	mcp.Tool,
	mcp.ToolHandlerFor[removeFactInput, removeFactOutput],
	error,
) {
	tool := mcp.Tool{
		Name:        "removeFactByID",
		Description: "Remove previously remembered fact/knowledge by ID",
	}

	var handler mcp.ToolHandlerFor[removeFactInput, removeFactOutput] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input removeFactInput,
	) (
		*mcp.CallToolResult,
		removeFactOutput,
		error,
	) {
		err := r.RemoveFactByID(ctx, input.ID)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, removeFactOutput{Ok: false}, err
		}

		return nil, removeFactOutput{Ok: true}, nil
	}

	return tool, handler, nil
}
