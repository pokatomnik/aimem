package web

import (
	"net/http"
	"strings"
)

// apiPrefix marks the requests owned by the generated server. Everything else
// is the SPA, because both are served from one binary on one port.
const apiPrefix = "/api/"

// newHandler routes between the API and the UI.
//
// The split lives here rather than in server.WithNotFound on purpose: a typo
// inside the prefix, such as /api/v2/search, then still reaches the API and is
// answered with a JSON 404, instead of quietly returning index.html and making
// the mistake hard to spot.
func newHandler(api, ui http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiPrefix) {
			api.ServeHTTP(w, r)

			return
		}

		ui.ServeHTTP(w, r)
	})
}
