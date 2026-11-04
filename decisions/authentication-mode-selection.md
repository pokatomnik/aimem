# Authentication Mode Selection

## Decision

Authentication follows the storage backend, not the contents of the config.
`users.RequireAuthentication` makes that one decision:

- embedded backend → `false`; every request runs anonymously, and
  `AIMEM_USERS_JSON`, if set, is ignored, because `kv` ignores `userID`;
- Qdrant → `true`, and startup **fails** with `ErrNoUsersConfigured` when the list
  defines no token.

It is asked once, in `scaffold`, with the context `main` provides. The answer is
expressed the same way for both servers: **two constructors each**, never a bool
argument and never a per-call check.

- `cmdWebPackage.NewWithAuth` / `NewWithoutAuth` set the unexported
  `authRequired` field.
- `aMCPPackage.NewWithAuth` / `NewWithoutAuth` inject either
  `tools.NewTokenAuthenticator(users)` or `tools.NewAnonymousAuthenticator()`.

There is one MCP package — the server, its tools and the streamable-HTTP transport
together — and each constructor returns a plain `http.Handler`, so the SDK's
`*mcp.Server` never reaches `scaffold`.

A tool calls `a.UserID(ctx, req)` and nothing else: which authenticator was
injected *is* the mode, so a tool holds no branch of its own and cannot disagree
with the web API.

The answer travels as `Methods.authRequired`, which `Methods.getUserID` enforces by
refusing an empty user. `getUserID` checks the mode **before** reading the
context, so with auth off the user is never looked up and a stale value cannot
travel into a store that ignores it.

## Why

The decision has to exist before the first request: a server that starts and then
answers 401 forever gives the operator no way out.

Two constructors make the mode the shape of the object instead of an argument every
caller has to get right; a `bool` at the end of a long argument list is exactly the
kind of thing that gets flipped by accident.

Deciding once in `scaffold` is what keeps the API and the tools from ending up in
different modes — one branch picks both from the same value. Credential parsing
(`tools.NewTokenAuthenticator`) stays apart from that decision, because turning a
header into a user ID is not the same job as judging a deployment.

One MCP package rather than a server package plus an "http" one: the second would
only build the first, and would misdescribe itself while the real HTTP server lives
in `internal/controllers/web`.
