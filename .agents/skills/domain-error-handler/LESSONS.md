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
