package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	getFactInput struct {
		ID string `json:"id" jsonschema:"ID of the fact to read, as returned by the recall tool"`
	}
	getFactOutput struct {
		ID        string `json:"id" jsonschema:"Fact/knowledge ID"`
		CreatedAt string `json:"createdAt" jsonschema:"When fact/knowledge was remembered"`
		Summary   string `json:"summary" jsonschema:"Short summary, the same text the recall tool shows"`
		Fact      string `json:"fact" jsonschema:"Full stored fact/knowledge contents"`
	}
)

func ToolGetFact(a mcpdeps.Authenticator, r mcpdeps.GetFact, l mcpdeps.Logger) (
	mcp.Tool,
	mcp.ToolHandlerFor[getFactInput, getFactOutput],
	error,
) {
	tool := mcp.Tool{
		Name: "read_fact",
		Description: "Read one saved fact/knowledge in full by its ID. The recall tool returns only short " +
			"summaries; call this with an ID from a recall result whenever the summary is not detailed enough " +
			"to answer — exact wording, names, numbers, paths, or anything the summary may have dropped. " +
			"The result contains the complete stored text plus the summary.",
	}

	var handler mcp.ToolHandlerFor[getFactInput, getFactOutput] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input getFactInput,
	) (
		*mcp.CallToolResult,
		getFactOutput,
		error,
	) {
		userID, err := a.UserID(ctx, req)
		if err != nil {
			logToolError(l, tool.Name, "", err)

			return &mcp.CallToolResult{IsError: true}, getFactOutput{}, err
		}

		fact, err := r.GetFact(ctx, userID, input.ID)
		if err != nil {
			logToolError(l, tool.Name, userID, err)

			return &mcp.CallToolResult{IsError: true}, getFactOutput{}, err
		}

		return nil, getFactOutput{
			ID:        fact.ID,
			CreatedAt: fact.CreatedAt,
			Summary:   fact.Summary,
			Fact:      fact.Content,
		}, nil
	}

	return tool, handler, nil
}
