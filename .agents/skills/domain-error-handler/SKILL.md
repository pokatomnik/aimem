---
name: domain-error-handler
description: Audit that every domain error in internal/entities/apperrors/errors.go is mapped in the NewError switch of internal/controllers/web/error_handler.go, that each mapping is correct (errors.Is, status code, errorCode enum value), and write the missing or wrong handlers. Records every user correction in LESSONS.md so the same mistake is not repeated. Use when domain errors are added or renamed, when the web error handler is edited, or when an HTTP response carries a wrong status or error code.
---

# Domain Error Handler

Keep the mapping from domain errors to HTTP responses complete and correct, and
learn from every correction the user makes.

The contract spans these files:

| File                                                 | Role                                                          |
| ---------------------------------------------------- | ------------------------------------------------------------- |
| `internal/entities/apperrors/errors.go`              | the list of domain errors (source of truth)                   |
| `internal/controllers/web/error_handler.go`          | `NewError`, which maps them to status codes                   |
| `internal/controllers/web/openapi.yaml`              | the `errorCode` enum, i.e. the codes allowed in a body        |
| `internal/controllers/web/server/oas_schemas_gen.go` | generated `ErrorCode*` constants — the only names you may use |

## Read LESSONS.md first

`LESSONS.md` in this skill's directory holds corrections the user has already
made. Read it before proposing any change.

- If a lesson applies, follow it and say which one you followed.
- If the code contradicts a lesson, stop and report the conflict. Do not pick a
  side on your own — a lesson records a decision the user already made once.
- Never treat this file as optional reading because "the case looks simple". It
  exists precisely for the cases that look simple.

## Order of work

1. Read `LESSONS.md`.
2. Extract the domain errors from `errors.go`.
3. Extract the cases from `NewError`.
4. Diff both directions — see below.
5. **Write the missing handlers immediately.** A domain error with no `case` is
   unambiguous debt; the user asked for it to be closed in the same run, not
   queued for approval. Add them, then report what you added.
6. **Report, and wait, for anything judgmental** — a mapping you want to change,
   a code you want to delete, a status you would reassign. These are the user's
   decisions; the code is written by hand and may be deliberate.
7. Verify.
8. If the user corrects you, fix the code **and** append a lesson.

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
  for, 5xx for a failure on our side (storage broken, model backend down).
- **`NewError` must always return a non-nil value with `StatusCode` set.**
  `encodeErrorResponse` dereferences `response.StatusCode`
  (`oas_response_encoders_gen.go`), so a nil return panics inside the handler
  instead of producing a response. A non-nil value with a zero status is not an
  error either: the encoder rewrites `0` into **200 OK**, so the failure travels
  as a success. `UnimplementedHandler.NewError` returns exactly such an empty
  value — never copy it.

### Direction 5 — the default is honest

The `default` branch catches every error nobody anticipated. Whatever status
and code it uses will be attached to an unknown failure, so it must not claim a
specific known cause. If no enum value means "unmapped failure", the spec needs
one (`INTERNAL_ERROR` already exists) — propose that instead of borrowing
another code.

## Adding a handler

1. Check whether an enum value already describes the cause. Reuse it only if it
   is the same meaning, not merely the same area.
2. If no value fits, add a SCREAMING_SNAKE entry to `errorCode` in
   `internal/controllers/web/openapi.yaml`.
3. Run `make generate`. This regenerates **both** `internal/controllers/web/server`
   and `cmd/api/server`, because `cmd/api/openapi.go` carries its own
   `go:generate` line. `--clean` wipes the target directory, so never put
   hand-written code inside a `server/` directory.
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

`go build` fails while `*Methods` misses a handler method ogen requires; that is
an unrelated, pre-existing failure. Confirm your file is not named in the error
before reporting success. If the terminal sandbox refuses to write
`~/Library/Caches/go-build`, request write access to `GOCACHE` and `GOMODCACHE`
instead of concluding the code is broken.

## Learning from corrections

This is the part that makes the skill improve. It runs only when the **user**
tells you a decision was wrong.

When that happens, do all three:

1. Fix the incorrect handler.
2. Append an entry to `LESSONS.md` — never rewrite or delete an existing one.
3. Keep the entry short enough to be read on every future run.

An entry must contain, in this order:

- **What I did** — the concrete mapping or edit, with the file.
- **What the user said** — the correction in their words, not your summary of it.
- **Why it was wrong** — the mechanism, not the verdict. "Status 400 hid a
  server bug behind a client error" is usable; "the status was wrong" is not.
- **Rule** — what to do instead, phrased so it can be checked mechanically.
- **Check** — the grep, build, or command that proves the rule holds.

If a new lesson contradicts an older one, add the new entry and mark the old one
`(superseded by #N)`. History stays visible; that is how a future run knows the
decision reversed and why.

Never write a lesson the user did not actually correct. Speculation belongs in
your reply, not in `LESSONS.md`; a false lesson is worse than no lesson,
because later runs will follow it.

## Reporting

Separate the findings:

- **Missing** — domain error with no `case`, with the file and line.
- **Wrong** — a case that matches incorrectly, groups distinct causes, or uses
  a status that contradicts the cause.
- **Dead code** — enum value nothing produces.
- **Judgment calls** — mappings that work but look misleading.

State the current debt every run instead of assuming it was already fixed.
