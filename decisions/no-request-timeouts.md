# No Request Timeouts

## Decision

Requests run without a deadline, so a slow backend blocks the CLI.

Cancellation exists only through signals: `main` wraps `context.Background()` in
`signal.NotifyContext` (SIGINT/SIGTERM/SIGQUIT) and prints `Cancelled` when the
context is canceled.

## Why

The expensive calls are the in-process embedding and the summarization request
to the configured model server: how long they take is a property of the machine
and the model, not of the request, so any deadline would be a guess that fails a
legitimate call on a slow host and then looks like a bug in the client.

The operator runs this CLI interactively, so the honest control is Ctrl-C: one
cancellation path that reaches every backend, instead of a per-request budget that
has to be chosen, enforced in every layer and documented.
