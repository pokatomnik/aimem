# One Listener, Three Routes

## Decision

`newHandler` in `internal/controllers/web/handler.go` splits requests by path
prefix, in this order:

1. `/mcp` → the MCP streamable-HTTP handler;
2. `/api/` → the generated API;
3. everything else → the SPA.

MCP is created once in `scaffold` and passed into the `web` controller as a plain
`http.Handler`, so the MCP endpoint shares the `web` port and authenticates every
API request and every MCP tool call with a bearer token.

## Why

The split lives in a handler rather than in `server.WithNotFound` so that a typo
inside the prefix — `/api/v2/search` — still reaches the API and is answered with a
JSON 404 instead of quietly returning `index.html`. A client that asked for an API
route gets an API answer, and the mistake stays visible instead of turning into a
page that parses as nothing.

Passing the endpoint as an `http.Handler` is what keeps the controller from naming
an adapter type (see `coupling-restrictions.md`), and building it in `scaffold` is
what lets one decision choose the authenticated or the anonymous MCP server.
