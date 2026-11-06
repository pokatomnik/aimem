# Logging At The Boundaries

## Decision

Logging is a cross-cutting concern owned by `internal/shared/logging`, which builds
a `log/slog` logger from `AIMEM_LOG_FILE` and `AIMEM_LOG_LEVEL`. The composition
root (`cmd/aimem/scaffold.go`) injects that logger into the web controller and the
MCP adapter through a narrow unexported `logger` interface (`Debug`, `Info`,
`Warn`, `Error`). With no `AIMEM_LOG_FILE` the log goes to stderr; with a path the
file is created if missing and opened for append.

Use cases and repositories do not hold a logger and do not log. They return
wrapped errors. A failure is logged once, at the edge that turns it into an HTTP
response or an MCP tool result: the web error handler, the MCP tools, and
startup/shutdown in `cmd/aimem`. A refusal caused by the caller or its data is a
`Warn`; a fault on our side is an `Error`.

## Why

The business logic stays free of a logger, so a use case can be read and tested
without arranging one, and no log entry is produced by a layer that does not know
whether the caller is HTTP, MCP, or a future transport. Logging inside a use case
would repeat the same failure at every layer it passes through and make the log's
volume depend on the call graph.

The alternative — a logger field in every struct that can fail — was rejected
because it puts a side effect in the middle of the operation and forces every test
of that operation to supply a logger to observe behaviour it is not testing.
