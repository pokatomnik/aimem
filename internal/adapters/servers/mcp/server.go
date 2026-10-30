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
func createServer(deps MCPDeps, a mcpdeps.Authenticator) (*sdk.Server, error) {
	implementation := &sdk.Implementation{
		Name:        "memory",
		Title:       "AI Memory Manager",
		Description: "Manages AI memories: remember and recall facts/knowledge/important information",
	}

	options := &sdk.ServerOptions{
		Instructions: instructions,
	}

	server := sdk.NewServer(implementation, options)

	toolRemember, rememberHandler, err := tools.ToolRemember(a, deps.Remember)
	if err != nil {
		return nil, err
	}

	toolRecall, recallHandler, err := tools.ToolRecall(a, deps.Recall)
	if err != nil {
		return nil, err
	}

	toolGetFact, getFactHandler, err := tools.ToolGetFact(a, deps.GetFact)
	if err != nil {
		return nil, err
	}

	toolRemoveFact, removeFactHandler, err := tools.ToolRemoveFact(a, deps.RemoveFact)
	if err != nil {
		return nil, err
	}

	sdk.AddTool(server, &toolRemember, rememberHandler)
	sdk.AddTool(server, &toolRecall, recallHandler)
	sdk.AddTool(server, &toolGetFact, getFactHandler)
	sdk.AddTool(server, &toolRemoveFact, removeFactHandler)

	return server, nil
}
