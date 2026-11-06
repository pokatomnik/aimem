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
		Summary   string `json:"summary" jsonschema:"A short summary, not the full text. Read the full fact with the read_fact tool"`
	}
	recallOutput struct {
		Facts []fact `json:"facts" jsonschema:"Facts/knowledge we have"`
	}
)

func ToolRecall(a mcpdeps.Authenticator, r mcpdeps.Recall, l mcpdeps.Logger) (
	mcp.Tool,
	mcp.ToolHandlerFor[recallInput, recallOutput],
	error,
) {
	tool := mcp.Tool{
		Name: "recall",
		Description: "Semantic search over saved facts/knowledge. Returns only the top-K most relevant matches, ranked " +
			"by similarity, each with an ID, creation time, and a short summary — never the full text. Because it is a " +
			"relevance search rather than a listing, an empty or short result does not mean nothing is stored, and it " +
			"cannot enumerate everything that is remembered. When a summary is too short to answer from, call " +
			"read_fact with that ID.",
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
		userID, err := a.UserID(ctx, req)
		if err != nil {
			logToolError(l, tool.Name, "", err)

			return &mcp.CallToolResult{IsError: true}, recallOutput{}, err
		}

		facts, err := r.Recall(ctx, userID, input.Query)
		if err != nil {
			logToolError(l, tool.Name, userID, err)

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
