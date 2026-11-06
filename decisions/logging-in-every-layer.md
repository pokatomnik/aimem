# Logging In Every Layer

## Decision

`scaffold` builds one `*slog.Logger` from `LOG_FILE` and `LOG_LEVEL` and injects
it as `internal/shared/logging.Logger` into adapters, repositories,
use cases, the web controller and its handlers, and the MCP tools. Every component
logs the error it is about to return, on a single source line, and then returns
the same error and values it returned before. Logging never changes control flow,
wrapping, or a business result. The logger is required: no component substitutes
a discard logger for nil, so a missed wiring fails loudly instead of silently
dropping records. Tests pass `logging.Discard()` explicitly.

A refusal caused by the caller or its data is `Warn`; a fault on our side is
`Error`. Startup and shutdown are `Info`. When `LOG_FILE` is empty the log goes
to stderr.

Every record answers three questions: that something failed (the message), where
it failed (the automatic `source` file and line, plus the operation named in the
message and its `tool` or `operation` attribute), and why it failed (the wrapped
`error` plus the identifiers the operation had — `user_id`, `fact_id`, `top_k`,
`status`, `path`, and so on). Credentials and user content are never logged;
lengths and identifiers are.

This supersedes `logging-at-boundaries.md`, which kept use cases and repositories
logger-free. Boundary-only logging left a failure without the operation that
produced it, and the project now wants that layer in the log.

## Why

The single-line rule keeps the log call from hiding the business statement it
guards, and returning the original value makes the change auditable: every test
that asserted the old behaviour still passes.

Logging per layer costs duplicate entries for one failure, but it gives the
operator the layer, the operation and the wrapped cause, which a boundary-only
message cannot. The alternative — decorators around the use cases — was rejected
because it would leave repositories and adapters silent, which is exactly the gap
this change closes.
