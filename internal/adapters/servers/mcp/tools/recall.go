package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type (
	recallInput struct {
		Query string `json:"query" jsonschema:"Search query"`
	}
	fact struct {
		ID        string `json:"id" jsonschema:"Fact/knowledge ID"`
		CreatedAt string `json:"createdAt" jsonschema:"Date when knowledge was created"`
		Summary   string `json:"summary" jsonschema:"A short fact/knowledge summary. Read full knowledge using \"getFactByID\" tool"`
	}
	recallOutput struct {
		Facts []fact `json:"facts" jsonschema:"Facts/knowledge we have"`
	}
)

func ToolRecall(r mcpdeps.Recall) (
	mcp.Tool,
	mcp.ToolHandlerFor[recallInput, recallOutput],
	error,
) {
	tool := mcp.Tool{
		Name:        "recall",
		Description: "Recall all related facts/knowledge",
	}

	var handler mcp.ToolHandlerFor[recallInput, recallOutput] = func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input recallInput,
	) (
		*mcp.CallToolResult,
		recallOutput,
		error,
	) {
		facts, err := r.Recall(ctx, input.Query)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, recallOutput{}, err
		}

		f := make([]fact, 0, len(facts))
		for _, cur := range facts {
			f = append(f, fact{
				ID:        cur.ID,
				CreatedAt: cur.CreatedAt,
				Summary:   cur.Summary,
			})
		}

		return nil, recallOutput{
			Facts: f,
		}, nil
	}

	return tool, handler, nil
}
