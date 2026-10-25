package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordingHandler marks which branch was taken by answering with its own
// status, so the test needs no knowledge of either side's real behaviour.
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

func TestNewHandlerRoutesByPrefix(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantAPI bool
	}{
		{
			name:    "spec path",
			target:  "/api/v1/search",
			wantAPI: true,
		},
		{
			name:    "spec path with an id",
			target:  "/api/v1/fact/abc123",
			wantAPI: true,
		},
		{
			name:    "unknown path inside the prefix stays on the API",
			target:  "/api/v2/search",
			wantAPI: true,
		},
		{
			name:    "root is the SPA",
			target:  "/",
			wantAPI: false,
		},
		{
			name:    "index is the SPA",
			target:  "/index.html",
			wantAPI: false,
		},
		{
			name:    "hashed asset is the SPA",
			target:  "/assets/index-CT2NxLkI.js",
			wantAPI: false,
		},
		{
			name:    "prefix without the trailing slash is the SPA",
			target:  "/api",
			wantAPI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &recordingHandler{status: http.StatusTeapot}
			ui := &recordingHandler{status: http.StatusOK}

			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.target, nil)

			newHandler(api, ui).ServeHTTP(rec, r)

			switch {
			case tt.wantAPI && !api.called:
				t.Fatal("API handler was not called")
			case !tt.wantAPI && !ui.called:
				t.Fatal("UI handler was not called")
			}

			if api.called && ui.called {
				t.Fatal("both handlers were called, the split is ambiguous")
			}

			wantStatus := ui.status
			if tt.wantAPI {
				wantStatus = api.status
			}

			if rec.Code != wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, wantStatus)
			}
		})
	}
}

// The request must reach the inner handler untouched, otherwise path, query and
// body would be lost at the split.
func TestNewHandlerPassesRequestThrough(t *testing.T) {
	api := &recordingHandler{status: http.StatusOK}
	ui := &recordingHandler{status: http.StatusOK}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/search?q=x", nil)

	newHandler(api, ui).ServeHTTP(httptest.NewRecorder(), r)

	if api.gotReq != r {
		t.Fatal("the API handler received a different *http.Request")
	}

	if got, want := api.gotReq.URL.RawQuery, "q=x"; got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}
