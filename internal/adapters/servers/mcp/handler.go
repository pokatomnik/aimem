package mcp

import (
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/tools"
)

// Server is this adapter's MCP server, ready to be mounted on an HTTP listener.
//
// The streamable HTTP transport is an implementation detail of the adapter: a
// caller mounts the value at a path prefix and reaches it as a plain http.Handler,
// so the SDK's *sdk.Server never leaks into the composition root. The transport is
// HTTP because the bearer token arrives in a header — it is the only transport that
// can carry the credentials this adapter is built around.
type Server struct {
	// server is the SDK server behind handler. It is kept so a test in this package
	// can drive it over the in-memory transport; nothing outside the package reaches
	// it, which is what keeps the transport an internal detail.
	server  *sdk.Server
	handler http.Handler
}

// NewWithAuth builds the server for a deployment where every tool call must carry
// a valid token.
func NewWithAuth(deps MCPDeps, authentication mcpdeps.Authentication) (*Server, error) {
	return newServer(deps, tools.NewTokenAuthenticator(authentication))
}

// NewWithoutAuth builds the server for the embedded backend, which stores one
// operator's local files and ignores the user ID. Tool calls run anonymously: no
// header is read and none is required.
func NewWithoutAuth(deps MCPDeps) (*Server, error) {
	return newServer(deps, tools.NewAnonymousAuthenticator())
}

// newServer is the single place the two constructors share, so the fields they
// set cannot drift apart.
func newServer(deps MCPDeps, a mcpdeps.Authenticator) (*Server, error) {
	srv, err := createServer(deps, a)
	if err != nil {
		return nil, err
	}

	return &Server{
		server:  srv,
		handler: newStreamableHandler(srv),
	}, nil
}

func newStreamableHandler(srv *sdk.Server) http.Handler {
	return sdk.NewStreamableHTTPHandler(func(_ *http.Request) *sdk.Server {
		return srv
	}, &sdk.StreamableHTTPOptions{
		JSONResponse:                 true,
		PropagateRequestCancellation: true,
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
