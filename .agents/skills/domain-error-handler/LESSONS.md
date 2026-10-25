# Lessons

Corrections the user has made to this skill's output. Read before every run.

Rules for this file:

- Entries are appended, never edited in place. A reversed decision gets a new
  entry and the old one is marked `(superseded by #N)`.
- Only real corrections by the user belong here. No speculation, no notes about
  what seemed odd.
- Each entry states what was done, what the user said, why it was wrong, the
  rule, and a check that proves the rule.

---

## 1. An unknown route is 404, not 400

**What I did.** In `internal/controllers/web/error_response.go`, `NotFoundHandler`
answered requests for paths absent from the spec with `400 BAD_REQUEST`, reasoning
that "the request does not match the contract".

**What the user said.** "Нет, ты тут тоже прав, нужен отдельно 404" — a missing
route needs its own 404.

**Why it was wrong.** A request that names no existing resource is not a
malformed request. 400 tells the client "your message is broken, fix it and
retry", which is useless when the address simply does not exist; 404 tells the
client the address is not served here. Collapsing both into 400 also makes the
two failures indistinguishable in logs and in the browser devtools.

Since `error` is a required field of `appError`, a 404 body needs a code that
matches it — `NOT_FOUND`, not `BAD_REQUEST`. A status and a body code that
disagree are worse than either alone.

**Rule.** Status codes in `error_response.go` and `error_handler.go` must
describe the actual cause:

- 400 — the body or parameters are malformed or fail spec validation.
- 404 — the path or the referenced entity does not exist.
- 4xx otherwise — the caller asked for something impossible or unpermitted.
- 5xx — our side failed; never report these as 4xx.

The code in the body must be the enum value that names the same cause; never
reuse an enum value of a different meaning just because a value exists.

**Check.**

```sh
grep -n "NotFoundHandler" -A 3 internal/controllers/web/error_response.go
```

Expect `http.StatusNotFound` together with `server.ErrorCodeNOTFOUND`.

---

## 2. Error handler errors must keep their domain status

**What I did.** `RequestErrorHandler` in `internal/controllers/web/error_response.go`
classified only decode and content-type errors as 400 and reported every other
error as `500 INTERNAL_ERROR`, while `Methods.NewError` held the real domain
mapping. When the generated server stopped calling `NewError` and routed handler
errors to `cfg.ErrorHandler` instead, `ErrFactDoesNotExist` became 500 and
`ErrAlreadyHasFact` stopped being 409.

**What the user said.** "Разрешаю, это большая проблема. Так быть не должно."

**Why it was wrong.** Two entry points turned errors into responses, but only
one of them knew the mapping. The generated handler path silently lost it, so a
client error was reported as a server fault — a missing fact looked like a
broken server, and clients lost the branches their spec promised. The tests
stayed green because `error_handler_test.go` exercised `NewError` directly, a
code path the server no longer calls.

**Rule.** The domain mapping must live in exactly one function,
`mapDomainError`, and every path that converts an error into a response must go
through it: `Methods.NewError` and `RequestErrorHandler` alike.
`RequestErrorHandler` checks `isRequestError` first (client fault → 400) and
delegates everything else to `mapDomainError`. Never hardcode a status for the
remaining branch.

**Check.**

```sh
CGO_ENABLED=0 go test ./internal/controllers/web/ -run 'TestRequestErrorHandlerDomainError|TestRequestErrorHandlerWrappedDomainError' -v
```

Expect 404/409/422/502 per domain error, including wrapped ones, and
`FACT_DOES_NOT_EXIST` rather than `INTERNAL_ERROR` in the body.
