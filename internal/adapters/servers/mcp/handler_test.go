package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/shared/server"
)

// inProcessTransport answers every request by invoking an http.Handler directly.
//
// It is what lets the streamable HTTP transport be driven end to end without a
// listener: httptest.NewServer would bind a loopback socket, and the transport's
// only requirement here is an http.Client. headers stand in for the client side of
// a deployment, where the token travels in a header.
type inProcessTransport struct {
	handler http.Handler
	headers http.Header
}

func (t *inProcessTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Path = "/mcp"

	for key, values := range t.headers {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}

	rec := httptest.NewRecorder()
	t.handler.ServeHTTP(rec, r)

	return rec.Result(), nil
}

// connectOver drives the adapter's own handler, i.e. the same value scaffold hands
// to the web controller. Going through the handler rather than the bare SDK server
// is the point: it proves the transport is wired, not just that a server exists.
func connectOver(t *testing.T, srv *Server, headers http.Header) *sdk.ClientSession {
	t.Helper()

	transport := &sdk.StreamableClientTransport{
		Endpoint:             "http://memory/mcp",
		HTTPClient:           &http.Client{Transport: &inProcessTransport{handler: srv, headers: headers}},
		DisableStandaloneSSE: true,
	}

	ctx := context.Background()

	client := sdk.NewClient(&sdk.Implementation{Name: "test"}, nil)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	t.Cleanup(func() { cs.Close() })

	return cs
}

// The token travels in an HTTP header, so the header has to survive the transport
// and land in the tool. This is the reason the MCP adapter wraps its own handler
// instead of exposing the SDK server: the transport is part of the wiring, not a
// caller's concern.
func TestHandlerPassesTheTokenToTheTool(t *testing.T) {
	f := &fakeDeps{}

	srv, err := NewWithAuth(MCPDeps{Remember: f, Recall: f, GetFact: f, RemoveFact: f}, fakeAuth{}, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	cs := connectOver(t, srv, http.Header{server.AuthorizationHeader: []string{"Bearer tok"}})

	res, err := cs.CallTool(context.Background(), &sdk.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": "anything"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatalf("authenticated call failed: %v", res.Content)
	}

	if want := "u-tok"; f.userID != want {
		t.Errorf("store received userID %q, want %q", f.userID, want)
	}
}

// Same transport, no header: the authenticated adapter must refuse, so a call
// cannot reach a store without a user.
func TestHandlerRefusesACallWithoutToken(t *testing.T) {
	f := &fakeDeps{}

	srv, err := NewWithAuth(MCPDeps{Remember: f, Recall: f, GetFact: f, RemoveFact: f}, fakeAuth{}, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	cs := connectOver(t, srv, nil)

	res, err := cs.CallTool(context.Background(), &sdk.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": "anything"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("call without a token was answered, want an error")
	}

	if f.userID != "" {
		t.Errorf("store was reached with userID %q, want no reach at all", f.userID)
	}
}

// The embedded wiring ignores the header entirely, so an anonymous call over the
// same transport must succeed.
func TestHandlerAcceptsACallWithoutTokenWhenAuthIsOff(t *testing.T) {
	f := &fakeDeps{}

	srv, err := NewWithoutAuth(MCPDeps{Remember: f, Recall: f, GetFact: f, RemoveFact: f}, testLogger())
	if err != nil {
		t.Fatal(err)
	}

	cs := connectOver(t, srv, nil)

	res, err := cs.CallTool(context.Background(), &sdk.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": "anything"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatalf("anonymous call failed: %v", res.Content)
	}
}
