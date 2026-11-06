# Log File Lifecycle

## Decision

`main` selects the log output: it passes `os.Stderr` when `LOG_FILE` is empty, or opens the configured file in append mode and passes the resulting `*os.File` to `scaffold`. `scaffold` creates the logger using that writer and `LOG_LEVEL`, then injects it into the dependency graph. `scaffold` neither opens nor closes the file; `main` closes a file it opened after command execution.

This supersedes the logger-ownership statement in [logging-in-every-layer.md](logging-in-every-layer.md); the logging rules in that record otherwise remain in force.

## Why

`main` owns process-level resources and their lifetime, while `scaffold` owns dependency construction. Passing the already-open writer keeps file I/O out of the composition root without making `main` assemble the dependency graph.
