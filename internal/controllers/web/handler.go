package web

import (
	"net/http"
	"strings"
)

const (
	apiPrefix = "/api/"
	mcpPrefix = "/mcp"
)

func newHandler(apiHandler, uiHandler, mcpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, mcpPrefix):
			mcpHandler.ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, apiPrefix):
			apiHandler.ServeHTTP(w, r)
		default:
			uiHandler.ServeHTTP(w, r)
		}
	})
}
