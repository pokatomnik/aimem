package stdio

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aimemMCP "github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
)

type StdIO struct {
	deps aimemMCP.MCPDeps
}

func New(deps aimemMCP.MCPDeps) *StdIO {
	return &StdIO{deps}
}

func (s *StdIO) Run(ctx context.Context) error {
	srv, err := aimemMCP.CreateServer(s.deps)
	if err != nil {
		return err
	}

	return srv.Run(ctx, &mcp.StdioTransport{})
}
