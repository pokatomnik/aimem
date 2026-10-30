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

// CreateServerWithAuth builds the MCP server for a deployment where every tool
// call must carry a valid token.
func CreateServerWithAuth(deps MCPDeps, authentication mcpdeps.Authentication) (*mcp.Server, error) {
	return createServer(deps, tools.NewTokenAuthenticator(authentication))
}

// CreateServerWithoutAuth builds the MCP server for the embedded backend, which
// stores one operator's local files and ignores the user ID. Tool calls run
// anonymously: no header is read and none is required.
func CreateServerWithoutAuth(deps MCPDeps) (*mcp.Server, error) {
	return createServer(deps, tools.NewAnonymousAuthenticator())
}

// createServer is the single place the two constructors share, so a tool cannot be
// wired with one authentication mode in one call and the other somewhere else.
func createServer(deps MCPDeps, a mcpdeps.Authenticator) (*mcp.Server, error) {
	implementation := &mcp.Implementation{
		Name:        "memory",
		Title:       "AI Memory Manager",
		Description: "Manages AI memories: remember and recall facts/knowledge/important information",
	}

	options := &mcp.ServerOptions{
		Instructions: instructions,
	}

	server := mcp.NewServer(implementation, options)

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

	mcp.AddTool(server, &toolRemember, rememberHandler)
	mcp.AddTool(server, &toolRecall, recallHandler)
	mcp.AddTool(server, &toolGetFact, getFactHandler)
	mcp.AddTool(server, &toolRemoveFact, removeFactHandler)

	return server, nil
}
