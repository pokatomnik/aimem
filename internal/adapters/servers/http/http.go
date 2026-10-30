package http

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aimemMCP "github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
)

type HTTP struct {
	handler http.Handler
}

// TryCreateWithAuth builds the streamable-HTTP handler for a deployment where every
// tool call must carry a valid token.
func TryCreateWithAuth(deps aimemMCP.MCPDeps, authentication mcpdeps.Authentication) (*HTTP, error) {
	srv, err := aimemMCP.CreateServerWithAuth(deps, authentication)
	if err != nil {
		return nil, err
	}

	return &HTTP{newStreamableHandler(srv)}, nil
}

// TryCreateWithoutAuth builds the streamable-HTTP handler for the embedded backend,
// where tool calls run anonymously.
func TryCreateWithoutAuth(deps aimemMCP.MCPDeps) (*HTTP, error) {
	srv, err := aimemMCP.CreateServerWithoutAuth(deps)
	if err != nil {
		return nil, err
	}

	return &HTTP{newStreamableHandler(srv)}, nil
}

func newStreamableHandler(srv *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:                 true,
		PropagateRequestCancellation: true,
	})
}

func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
