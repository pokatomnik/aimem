---
name: domain-error-handler
description: Audit that every domain error in internal/entities/apperrors/errors.go is mapped by mapDomainError in internal/controllers/web/error_handler.go, that each mapping is correct (errors.Is, status code, errorCode enum value), and that every path turning an error into a response goes through that one mapping. Use when domain errors are added or renamed, when the web error handler is edited, or when an HTTP response carries a wrong status or error code.
---

# Domain Error Handler

Keep the mapping from domain errors to HTTP responses complete and correct.

The contract spans these files:

| File                                                 | Role                                                          |
| ---------------------------------------------------- | ------------------------------------------------------------- |
| `internal/entities/apperrors/errors.go`              | the list of domain errors (source of truth)                   |
| `internal/controllers/web/error_handler.go`          | `mapDomainError`, the one mapping to a status and a body code |
| `internal/controllers/web/error_response.go`         | `RequestErrorHandler` and `NotFoundHandler`, which call it    |
| `internal/controllers/web/openapi.yaml`              | the `errorCode` enum, i.e. the codes allowed in a body        |
| `internal/controllers/web/server/oas_schemas_gen.go` | generated `ErrorCode*` constants — the only names you may use |

## One mapping, reached from everywhere

`mapDomainError` is the only place an error becomes a status and a body code,
and both entry points go through it:

- `RequestErrorHandler` checks `isRequestError` first — a malformed body, a
  failed parameter decode or a bad Content-Type is a client fault, so it becomes
  400 — and hands everything else to `mapDomainError`.
- Errors returned by our handlers reach `RequestErrorHandler` too, because the
  generated server routes them to the configured error handler instead of a
  per-operation `NewError` method.

So a domain error keeps its status on both paths: a missing fact stays 404, a
duplicate stays 409. Never add a second mapping, and never hardcode a status for
the remaining branch — two entry points where only one knows the mapping is how
`ErrFactDoesNotExist` once became 500 while the tests stayed green, because those
tests exercised a path the server no longer called.

```sh
CGO_ENABLED=0 go test ./internal/controllers/web/ \
  -run 'TestRequestErrorHandlerDomainError|TestRequestErrorHandlerWrappedDomainError' -v
```

Expect 404/409/422/502 per domain error, including wrapped ones, and
`FACT_DOES_NOT_EXIST` rather than `INTERNAL_ERROR` in the body.

## Order of work

1. Extract the domain errors from `errors.go`.
2. Extract the cases from `mapDomainError`.
3. Diff both directions — see below.
4. **Write the missing handlers immediately.** A domain error with no `case` is
   unambiguous debt; the user asked for it to be closed in the same run, not
   queued for approval. Add them, then report what you added.
5. **Report, and wait, for anything judgmental** — a mapping you want to change,
   a code you want to delete, a status you would reassign. These are the user's
   decisions; the code is written by hand and may be deliberate.
6. Verify.
7. If the user corrects you, fix the code and record the correction as described
   at the end of this file.

Adding is safe to do unprompted; changing or deleting is not. When in doubt,
which action you are taking decides whether to wait.

## The audit

### Direction 1 — every domain error is handled

```sh
grep -n "Err[A-Z]" internal/entities/apperrors/errors.go
grep -n "apperrors\.Err" internal/controllers/web/error_handler.go
```

Every `apperrors.Err*` must appear in a `case`. A domain error with no case
falls through to the default, which reports a generic 500 with a code borrowed
from another error — the caller gets a wrong code and no hint that the mapping
is missing.

### Direction 2 — every case refers to a real error

A case naming a deleted or renamed error does not compile. If the package still
builds, this direction is clean; if you find a case for an error that no longer
exists, remove it rather than renaming it into something plausible.

### Direction 3 — every enum value is reachable

```sh
grep -n "ErrorCode[A-Z]* ErrorCode = " internal/controllers/web/server/oas_schemas_gen.go
```

An enum value that no `case` produces is dead: the spec promises a code that
the server can never send, and a client written from the spec has a branch that
never fires. Report it. Whether to map a dead code to an existing error or to
delete it from `openapi.yaml` is the user's decision, not yours.

### Direction 4 — each check matches correctly

- **Use `errors.Is(err, apperrors.ErrX)`.** Repositories wrap errors with
  `fmt.Errorf("...: %w", ...)` (`internal/repositories/kv/getfact.go`,
  `getfacts.go`), so `err == apperrors.ErrX` and a tagged `switch err` both miss
  wrapped values. The `switch` must stay expressionless.
- **Group only errors that mean the same thing.** Two errors sharing one `case`
  send one code for two different causes; the client cannot tell them apart.
  Grouping is legitimate when the response really is identical, and wrong when
  it erases a distinction the caller needs.
- **Status code follows the cause.** 4xx for something the caller did or asked
  for, 5xx for a failure on our side (storage broken, model backend down). A path
  the spec does not serve is 404, not 400: `NotFoundHandler` answers it, and a 400
  would tell the client to fix a request that is not broken, besides making the
  two failures indistinguishable in logs and devtools.
- **The body code names the same cause as the status.** A 404 body carries
  `NOT_FOUND`, never `BAD_REQUEST` borrowed because the value exists. A status and
  a code that disagree are worse than either alone.
- **Never emit a zero status.** The per-operation encoders (`encodeAddFactResponse`
  and friends in `oas_response_encoders_gen.go`) rewrite an `AppErrorStatusCode`
  with `StatusCode == 0` into **200 OK**, so a failure would travel as a success.
  Always set the status explicitly.

### Direction 5 — the default is honest

The fallthrough at the end of `mapDomainError` catches every error nobody
anticipated. Whatever status and code it uses will be attached to an unknown
failure, so it must not claim a specific known cause. If no enum value means
"unmapped failure", the spec needs one (`INTERNAL_ERROR` already exists) — propose
that instead of borrowing another code.

## Adding a handler

1. Check whether an enum value already describes the cause. Reuse it only if it
   is the same meaning, not merely the same area.
2. If no value fits, add a SCREAMING_SNAKE entry to `errorCode` in
   `internal/controllers/web/openapi.yaml`.
3. Run `make api-generate`. It regenerates `internal/controllers/web/server` from
   `openapi.yaml` and the UI client types from the same spec — an enum entry is
   part of the contract, so both sides have to move together. `--clean` wipes the
   target directory, so never put hand-written code inside a `server/` directory.
4. Take the constant name from `oas_schemas_gen.go`, not from the spec. ogen
   strips underscores: `FACT_DOES_NOT_EXIST` → `ErrorCodeFACTDOESNOTEXIST`,
   `BAD_REQUEST` → `ErrorCodeBADREQUEST`. A guessed name does not compile.
5. Add the `case` with `errors.Is`, the status code, and that constant.

## Verify

```sh
gofmt -l internal/controllers/web/
go build ./internal/controllers/web/
CGO_ENABLED=0 go test ./...
```

`go build ./internal/controllers/web/` must be clean: a handler method ogen
requires but `*Methods` does not implement fails the build. If the terminal
sandbox refuses to write `~/Library/Caches/go-build`, request write access to
`GOCACHE` and `GOMODCACHE` instead of concluding the code is broken.

## Recording a correction

This is the part that makes the skill improve. It runs only when the **user**
tells you a decision was wrong — not when the code merely looks suspicious.

1. Fix the incorrect mapping.
2. Write the rule down where the next run will read it:

   - a rule about the mapping → the audit bullet or section it governs in this
     file, one line, with the reason in a clause;
   - a decision about the project → the matching file in `decisions/`, rewritten
     in the `## Decision` / `## Why` format described in `AGENTS.md`.

Rewrite the section instead of appending a story. Every run reads this whole file,
so it has to stay instructions, and git already keeps the history. Never record a
rule without its reason: "400 hid a server bug behind a client error" is a reason;
"the status was wrong" is not.

Never record a rule the user did not actually correct. Speculation belongs in your
reply, and a false rule is worse than none, because later runs will follow it.

## Reporting

Separate the findings:

- **Missing** — domain error with no `case`, with the file and line.
- **Wrong** — a case that matches incorrectly, groups distinct causes, or uses
  a status that contradicts the cause.
- **Dead code** — enum value nothing produces.
- **Judgment calls** — mappings that work but look misleading.

State the current debt every run instead of assuming it was already fixed.
