package mcp

import (
	_ "embed"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/tools"
)

//go:embed INSTRUCTIONS.md
var instructions string

type MCPDeps struct {
	Remember   mcpdeps.Remember
	Recall     mcpdeps.Recall
	GetFact    mcpdeps.GetFact
	RemoveFact mcpdeps.RemoveFact
}

// createServer is the single place the two constructors of Server share, so a tool
// cannot be wired with one authentication mode in one call and the other somewhere
// else.
func createServer(deps MCPDeps, a mcpdeps.Authenticator, l mcpdeps.Logger) (*sdk.Server, error) {
	implementation := &sdk.Implementation{
		Name:        "memory",
		Title:       "AI Memory Manager",
		Description: "Manages AI memories: remember and recall facts/knowledge/important information",
	}

	options := &sdk.ServerOptions{
		Instructions: instructions,
	}

	server := sdk.NewServer(implementation, options)

	toolRemember, rememberHandler, err := tools.ToolRemember(a, deps.Remember, l)
	if err != nil {
		l.Error("build remember tool failed", "tool", "remember", "error", err)

		return nil, err
	}

	toolRecall, recallHandler, err := tools.ToolRecall(a, deps.Recall, l)
	if err != nil {
		l.Error("build recall tool failed", "tool", "recall", "error", err)

		return nil, err
	}

	toolGetFact, getFactHandler, err := tools.ToolGetFact(a, deps.GetFact, l)
	if err != nil {
		l.Error("build read_fact tool failed", "tool", "read_fact", "error", err)

		return nil, err
	}

	toolRemoveFact, removeFactHandler, err := tools.ToolRemoveFact(a, deps.RemoveFact, l)
	if err != nil {
		l.Error("build remove_fact_by_id tool failed", "tool", "remove_fact_by_id", "error", err)

		return nil, err
	}

	sdk.AddTool(server, &toolRemember, rememberHandler)
	sdk.AddTool(server, &toolRecall, recallHandler)
	sdk.AddTool(server, &toolGetFact, getFactHandler)
	sdk.AddTool(server, &toolRemoveFact, removeFactHandler)

	return server, nil
}
