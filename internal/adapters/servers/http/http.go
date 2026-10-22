package http

import (
	"context"
	"errors"
	"net"
	stdHttp "net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aimemMCP "github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
	"github.com/pokatomnik/gomem/internal/shared/server"
)

type HTTP struct {
	deps aimemMCP.MCPDeps
}

func New(deps aimemMCP.MCPDeps) *HTTP {
	return &HTTP{deps}
}

func (h *HTTP) getHandler() (stdHttp.Handler, error) {
	srv, err := aimemMCP.CreateServer(h.deps)
	if err != nil {
		return nil, err
	}

	return mcp.NewStreamableHTTPHandler(func(r *stdHttp.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:                 true,
		PropagateRequestCancellation: true,
	}), nil
}

func (h *HTTP) Run(ctx context.Context, host string, options *server.ListenAndServeOptions) error {
	handler, err := h.getHandler()
	if err != nil {
		return err
	}

	httpServer := server.NewHTTPServerExt(&stdHttp.Server{
		Addr:    host,
		Handler: handler,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	})

	err = httpServer.ListenAndServeContext(ctx, options)
	if errors.Is(err, stdHttp.ErrServerClosed) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}
