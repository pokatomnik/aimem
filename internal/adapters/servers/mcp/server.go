package mcp

import (
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func CreateServer(deps MCPDeps) (*mcp.Server, error) {
	implementation := &mcp.Implementation{
		Name:        "memory",
		Title:       "AI Memory Manager",
		Description: "Manages AI memories: remember and recall facts/knowledge/important information",
	}

	options := &mcp.ServerOptions{
		Instructions: instructions,
	}

	server := mcp.NewServer(implementation, options)

	toolRemember, rememberHandler, err := tools.ToolRemember(deps.Remember)
	if err != nil {
		return nil, err
	}

	toolRecall, recallHandler, err := tools.ToolRecall(deps.Recall)
	if err != nil {
		return nil, err
	}

	toolGetFact, getFactHandler, err := tools.ToolGetFact(deps.GetFact)
	if err != nil {
		return nil, err
	}

	toolRemoveFact, removeFactHandler, err := tools.ToolRemoveFact(deps.RemoveFact)
	if err != nil {
		return nil, err
	}

	mcp.AddTool(server, &toolRemember, rememberHandler)
	mcp.AddTool(server, &toolRecall, recallHandler)
	mcp.AddTool(server, &toolGetFact, getFactHandler)
	mcp.AddTool(server, &toolRemoveFact, removeFactHandler)

	return server, nil
}
