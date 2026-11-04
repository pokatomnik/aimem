# The User List Is Parsed Once

## Decision

`internal/repositories/users` parses `AIMEM_USERS_JSON` — a JSON array of
`{"id": ..., "tokens": [...]}` entries — once, on first use, and caches it for the
lifetime of the process. A change takes effect on the next start.

A malformed list degrades to an empty map. A token listed under two users resolves
to the later one; duplicates are accepted.

The package is used only for Qdrant authentication and is ignored by the embedded
backend.

## Why

The list is one operator's hand-written config, and the auth mode is fixed before
the first request (see `authentication-mode-selection.md`), so re-reading it per
request would add I/O and a second source of truth that can change mid-flight.

Degrading to an empty map is deliberate: with Qdrant it surfaces as
`ErrNoUsersConfigured` and stops startup, so the deployment fails loudly instead of
accepting tokens nobody configured. Reporting the parse error itself would be a
logging concern, not a second error path out of `New`.

Duplicates are accepted because the list is small enough that a duplicate is
visible on sight, and "the later entry wins" is one sentence instead of a
validation step whose error nobody would read.
