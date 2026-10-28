package http

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aimemMCP "github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
)

type HTTP struct {
	handler http.Handler
}

func TryCreate(deps aimemMCP.MCPDeps) (*HTTP, error) {
	handler, err := getHandler(deps)
	if err != nil {
		return nil, err
	}

	return &HTTP{handler}, nil
}

func getHandler(deps aimemMCP.MCPDeps) (http.Handler, error) {
	srv, err := aimemMCP.CreateServer(deps)
	if err != nil {
		return nil, err
	}

	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:                 true,
		PropagateRequestCancellation: true,
	}), nil
}

func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
