package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	getFactInput struct {
		ID string `json:"id" jsonschema:"Fact/knowledge ID"`
	}
	getFactOutput struct {
		ID        string `json:"id" jsonschema:"Fact/knowledge ID"`
		CreatedAt string `json:"createdAt" jsonschema:"When fact/knowledge was remembered"`
		Summary   string `json:"summary" jsonschema:"Summary of the fact/knowledge"`
		Fact      string `json:"fact" jsonschema:"Fact/knowledge contents"`
	}
)

func ToolGetFact(r mcpdeps.GetFact) (
	mcp.Tool,
	mcp.ToolHandlerFor[getFactInput, getFactOutput],
	error,
) {
	tool := mcp.Tool{
		Name:        "get_fact_by_id",
		Description: "Get a fact/knowledge by ID",
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
		fact, err := r.GetFact(ctx, input.ID)
		if err != nil {
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
