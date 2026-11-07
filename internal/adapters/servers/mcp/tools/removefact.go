package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	removeFactInput struct {
		ID string `json:"id" jsonschema:"ID of the fact the user explicitly requested to delete, as returned by the recall tool"`
	}

	removeFactOutput struct {
		Ok bool `json:"ok" jsonschema:"True on successful completion, including when the fact was already absent"`
	}
)

func ToolRemoveFact(a mcpdeps.Authenticator, r mcpdeps.RemoveFact, l mcpdeps.Logger) (
	mcp.Tool,
	mcp.ToolHandlerFor[removeFactInput, removeFactOutput],
	error,
) {
	tool := mcp.Tool{
		Name: "remove_fact_by_id",
		Description: "Permanently delete by ID only on explicit user request. " +
			"Ask whether to remove conflicting facts. Returns success even if absent.",
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
		userID, err := a.UserID(ctx, req)
		if err != nil {
			logToolError(l, tool.Name, "", err)

			return &mcp.CallToolResult{IsError: true}, removeFactOutput{}, err
		}

		err = r.RemoveFactByID(ctx, userID, input.ID)
		if err != nil {
			logToolError(l, tool.Name, userID, err)

			return &mcp.CallToolResult{IsError: true}, removeFactOutput{Ok: false}, err
		}

		return nil, removeFactOutput{Ok: true}, nil
	}

	return tool, handler, nil
}
