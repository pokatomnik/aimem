package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordingHandler marks which branch was taken by answering with its own
// status, so the test needs no knowledge of any side's real behaviour.
type recordingHandler struct {
	status int
	called bool
	gotReq *http.Request
}

func (h *recordingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.gotReq = r
	w.WriteHeader(h.status)
}

// branchHandlers keeps one recorder per branch of the split, each answering
// with a status no other branch uses, so a wrong route is visible both in
// which handler ran and in the status that reached the client.
type branchHandlers struct {
	api *recordingHandler
	ui  *recordingHandler
	mcp *recordingHandler
}

func newBranchHandlers() *branchHandlers {
	return &branchHandlers{
		api: &recordingHandler{status: http.StatusTeapot},
		ui:  &recordingHandler{status: http.StatusOK},
		mcp: &recordingHandler{status: http.StatusAccepted},
	}
}

func (b *branchHandlers) handler() http.Handler {
	return newHandler(b.api, b.ui, b.mcp)
}

func (b *branchHandlers) byName(name string) *recordingHandler {
	switch name {
	case "api":
		return b.api
	case "ui":
		return b.ui
	case "mcp":
		return b.mcp
	default:
		panic("unknown branch: " + name)
	}
}

func (b *branchHandlers) calledCount() int {
	n := 0

	for _, h := range []*recordingHandler{b.api, b.ui, b.mcp} {
		if h.called {
			n++
		}
	}

	return n
}

func TestNewHandlerRoutesByPrefix(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "spec path",
			target: "/api/v1/search",
			want:   "api",
		},
		{
			name:   "spec path with an id",
			target: "/api/v1/fact/abc123",
			want:   "api",
		},
		{
			name:   "unknown path inside the prefix stays on the API",
			target: "/api/v2/search",
			want:   "api",
		},
		{
			name:   "MCP root",
			target: "/mcp",
			want:   "mcp",
		},
		{
			name:   "MCP path below the prefix",
			target: "/mcp/v1/tools",
			want:   "mcp",
		},
		{
			name:   "root is the SPA",
			target: "/",
			want:   "ui",
		},
		{
			name:   "index is the SPA",
			target: "/index.html",
			want:   "ui",
		},
		{
			name:   "hashed asset is the SPA",
			target: "/assets/index-CT2NxLkI.js",
			want:   "ui",
		},
		{
			name:   "prefix without the trailing slash is the SPA",
			target: "/api",
			want:   "ui",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := newBranchHandlers()

			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.target, nil)

			handlers.handler().ServeHTTP(rec, r)

			want := handlers.byName(tt.want)
			if !want.called {
				t.Fatalf("%s handler was not called", tt.want)
			}

			if handlers.calledCount() != 1 {
				t.Fatal("more than one handler was called, the split is ambiguous")
			}

			if rec.Code != want.status {
				t.Errorf("status = %d, want %d", rec.Code, want.status)
			}
		})
	}
}

// The request must reach the inner handler untouched, otherwise path, query and
// body would be lost at the split.
func TestNewHandlerPassesRequestThrough(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "API",
			target: "/api/v1/search?q=x",
			want:   "api",
		},
		{
			name:   "MCP",
			target: "/mcp?session=y",
			want:   "mcp",
		},
		{
			name:   "SPA",
			target: "/some/route?z=w",
			want:   "ui",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := newBranchHandlers()

			r := httptest.NewRequest(http.MethodPost, tt.target, nil)

			handlers.handler().ServeHTTP(httptest.NewRecorder(), r)

			got := handlers.byName(tt.want).gotReq
			if got == nil {
				t.Fatalf("%s handler was not called", tt.want)
			}

			if got != r {
				t.Fatal("the inner handler received a different *http.Request")
			}

			if want := r.URL.RawQuery; got.URL.RawQuery != want {
				t.Errorf("query = %q, want %q", got.URL.RawQuery, want)
			}
		})
	}
}
