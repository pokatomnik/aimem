---
name: storage-backend-parity
description: Audits that aimem's two vector-storage adapters — the embedded xvec repository in internal/repositories/kv and the Qdrant repository in internal/repositories/qdrant — satisfy the same interfaces the use cases declare and behave the same for add, get, list and remove, allowing only the documented differences (ID shape, per-user partition, score direction, duplicate-detection window, provisioning and handle lifecycle). Use when a storage operation, its errors, its schema, the ID derivation or the backend selection changes; when a fact behaves differently depending on whether AIMEM_QDRANT_HOST is set; when a third backend is added; or when asked whether the backends are in parity.
---

# Storage Backend Parity

One contract, two implementations. A use case — and therefore the web API and the
MCP tools — cannot tell which backend it was given, so everything a caller can
observe has to match. Everything that does not match has to be one of the accepted
differences recorded below.

This skill checks two things, in this order:

1. **One interface.** Both repositories satisfy the exact interfaces the use
   cases declare, and nothing outside `scaffold` knows which one is in use.
2. **One behaviour.** For the same input, both produce the same domain outcome:
   the same stored shape, the same sentinel errors, the same ranking and the same
   missing-fact answers.

## What is compared

| Role      | Embedded                                                                          | Qdrant                                        |
| --------- | --------------------------------------------------------------------------------- | --------------------------------------------- |
| Contract  | `internal/use_cases/*/use_case.go` — the four unexported `store` interfaces       | same                                          |
| Adapter   | `internal/repositories/kv`                                                        | `internal/repositories/qdrant`                |
| Selection | `cmd/aimem/scaffold.go` (`cfg.RequireAuth()`, i.e. non-empty `AIMEM_QDRANT_HOST`) | same                                          |
| Schema    | `kv/consts.go`, `kv.New`                                                          | `qdrant/consts.go`, `qdrant.createCollection` |
| Tests     | `kv/*_test.go`                                                                    | `qdrant/*_test.go`                            |

Read these decision records before judging anything; they are the reason the
accepted differences exist:

- `decisions/storage-backend-selection.md`
- `decisions/fact-id-from-text.md`
- `decisions/id-validity-is-backend-local.md`
- `decisions/delete-is-idempotent.md`
- `decisions/per-user-collections.md`
- `decisions/semantic-recall-only.md`
- `decisions/stored-records-immutable.md`
- `decisions/field-naming-per-layer.md`
- `decisions/coupling-restrictions.md`

## Check 1 — one interface

The contract belongs to the consumers, not to the repositories. Each use case
declares the single method it needs:

| Consumer                         | Method                                                                                                                                             |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `use_cases/remember` (`store`)   | `AddFact(ctx context.Context, userID string, vector []float32, summary, content string) error`                                                     |
| `use_cases/recall` (`store`)     | `GetFactsList(ctx context.Context, userID string, vector []float32, topK int, similarityThreshold float32) (facts []fact.FactListView, err error)` |
| `use_cases/getfact` (`store`)    | `GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error)`                                                               |
| `use_cases/removefact` (`store`) | `RemoveFactByID(ctx context.Context, userID, factID string) error`                                                                                 |

Go matches on types only, so a renamed parameter is not a compile error — but a
mismatch there almost always means one side was edited alone, so treat it as a
finding anyway.

```sh
# each exported method must appear exactly twice, once per backend
grep -rhoE 'func \([a-z]+ \*(KV|Qdrant)\) [A-Z][A-Za-z]*' internal/repositories \
  | sed -E 's/func \([a-z]+ \*(KV|Qdrant)\) //' | sort | uniq -c

# the contract as the consumers write it
grep -rn "AddFact\|GetFactByID\|GetFactsList\|RemoveFactByID" --include=*.go internal/use_cases/
```

Every count must be `2`. A count of `1` is a method only one backend has; zero is
a contract nobody satisfies. Then read the four use-case declarations and the two
implementations and compare their parameter and result types.

Rules:

- **No backend-only exported methods.** An exported method that only `*KV` or only
  `*Qdrant` has is a divergence even if no caller uses it yet: the next caller will
  build on one backend and the other will not compile. Internal helpers
  (`hash`, `pointExists`, `ensureCollection`, `getCollection`) are fine.
- **Both assert the contract at compile time.** `qdrant/contract_test.go` holds a
  `store` interface — a hand-written copy of the four use-case interfaces — and
  `var _ store = (*Qdrant)(nil)`. A copy can drift, so re-diff it against the use
  cases whenever a signature changes. `kv` currently has no such assertion; add
  one when it is missing (see `kv/getfact_test.go` for the fake it would sit next
  to). The assertion is what turns a signature drift into a build failure.
- **Only `scaffold` chooses.** Nothing outside `cmd/aimem` may name either
  repository package:

  ```sh
  grep -rln "internal/repositories/kv\|internal/repositories/qdrant" --include=*.go . \
    | grep -v _test.go
  ```

  Every result must be `cmd/aimem/scaffold.go` or a file inside the package being
  imported. A controller, adapter or use case naming a backend — or switching on
  its type — is a bug: the backend is chosen once, in one file.

- **Constructors may differ.** `kv.New(config, dataPath)` and
  `qdrant.New(config, *http.Client)` take different dependencies, and each package
  declares its own narrow `config`. That is construction, not contract, and it is
  expected (accepted difference 7).

## Check 2 — one behaviour

### The four operations

Both adapters must, for the same input:

**`AddFact`**

- derive the fact ID from `content` alone, with the same SHA-256 digest;
- store `created_at` (UTC, RFC3339), `summary`, `fact` and the vector under
  `embedding` — the same key set, in both;
- answer a text already stored by the same user with `ErrAlreadyHasFact` and never
  overwrite it;
- return any other failure as itself.

**`GetFactByID`**

- return `ID`, `CreatedAt`, `Summary`, `Content` with the same meaning;
- answer an unknown fact with `ErrFactDoesNotExist`;
- answer an ID the backend cannot store with `ErrFactDoesNotExist` too, never with
  `ErrDBConvertError` (`### IDs that cannot name a fact`);
- answer a stored field that is missing or not a string with `ErrDBConvertError`
  naming the key;
- read only `created_at`, `summary`, `fact` — never the vector;
- not reveal a fact held by another user (Qdrant: it is absent; `kv` has one store).

**`GetFactsList`**

- apply `topK` first, then filter by `similarity >= similarityThreshold` — a
  result set may be empty even when similar facts exist;
- return best match first;
- return `ID`, `CreatedAt`, `Summary` only;
- answer a field of the wrong type with `ErrDBConvertError`;
- answer `topK <= 0` with an error (the message differs, that is accepted);
- return an empty, non-nil slice when nothing matches.

**`RemoveFactByID`**

- delete only the caller's fact;
- **succeed** when the fact is not stored — `DELETE` is idempotent, and the only
  failures it may report are infrastructure ones, never "absent" (`### Deleting`);
- succeed for an ID the backend cannot store, without opening the store;
- never patch a stored record; delete-then-insert is the only way to change one.

### The stored shape

Diff these mechanically; a difference is a bug, not a difference of dialect:

| Aspect      | Embedded                                | Qdrant                           |
| ----------- | --------------------------------------- | -------------------------------- |
| Metric      | `xvec.MetricTypeCosine`                 | `distanceCosine = "Cosine"`      |
| Vector name | `keyFieldEmbedding = "embedding"`       | same                             |
| Datatype    | `xvec.DataTypeVectorFP32`               | `datatypeFloat32 = "float32"`    |
| Dimensions  | `config.GetDimensions()`                | same                             |
| Base name   | schema name `"memory"`                  | `collectionName` base `"memory"` |
| Field keys  | `created_at`, `summary`, `fact`         | same                             |
| Timestamp   | `time.Now().UTC().Format(time.RFC3339)` | same                             |

```sh
grep -n 'keyField[a-zA-Z]* *= *"' internal/repositories/kv/consts.go \
  internal/repositories/qdrant/consts.go
grep -n "MetricTypeCosine\|DataTypeVectorFP32\|Dimension\|NewCollectionSchema" \
  internal/repositories/kv/kv.go
grep -n "distanceCosine\|datatypeFloat32\|Distance\|Size:" \
  internal/repositories/qdrant/*.go
grep -rn "time.Now()" internal/repositories/kv internal/repositories/qdrant
```

`snake_case` in storage is fixed by `decisions/field-naming-per-layer.md`; the
naming greps live in `.agents/skills/architecture-audit` and are not repeated here.

### The sentinel errors

Three sentinels cross the contract. For each cause, both backends must produce the
same one:

| Cause                                     | Sentinel              |
| ----------------------------------------- | --------------------- |
| The text is already stored by this user   | `ErrAlreadyHasFact`   |
| The fact is not there, **on read**        | `ErrFactDoesNotExist` |
| A stored field is missing or not a string | `ErrDBConvertError`   |

```sh
grep -rn "apperrors.Err" internal/repositories/kv internal/repositories/qdrant \
  | grep -v _test.go
```

Ignore hits inside comments; compare call sites. Every other failure is
infrastructure, and its type is expected to differ (accepted difference 6). Never
wrap a transport error into a sentinel it does not mean: a dead Qdrant server must
not read as "fact does not exist", and an xvec I/O error must not read as a
duplicate.

### IDs that cannot name a fact

A read or a delete takes an ID the caller chose, and the two backends accept
different alphabets: Qdrant stores a point ID only as a UUID or a `u64`, while the
embedded store accepts any UTF-8 key of 1..`xvec.MaxPrimaryKeyBytes`. So each
adapter defines a set of strings that cannot name a fact in it, and that set has
nothing to do with the domain. The contract deliberately declares no `factID`
format, so this rule lives in `decisions/id-validity-is-backend-local.md`.

A string outside that set is answered like an absent fact — never as a storage
failure — and the check runs before the store is touched, so no collection is
opened and no request is sent. Qdrant detects one in `pointIDFromExternal`, `kv`
in `validFactID` (invalid UTF-8, or longer than `xvec.MaxPrimaryKeyBytes`).

| Operation        | Answer                                            |
| ---------------- | ------------------------------------------------- |
| `GetFactByID`    | `ErrFactDoesNotExist` → 404 `FACT_DOES_NOT_EXIST` |
| `RemoveFactByID` | success → 200, no-op                              |

**Do not audit the empty ID through HTTP.** ogen's parameter decoder treats an
empty `{factID}` segment as a missing required parameter and answers `400`
`BAD_REQUEST` before any handler runs, so `GET` and `DELETE /api/v1/fact/` never
reach either adapter. That `400` is settled and intended — it is not a divergence
to fix. The adapters still guard `""` because the MCP tools take `id` with no such
decoder in front of them; a tool call with an empty ID is what keeps that branch
exercised.

`ErrDBConvertError` is reserved for stored data: a field that is missing or not a
string. Using it for caller input reports our storage as corrupted, and because it
maps to `500` it once turned `GET /api/v1/fact/abc` into 404 on the embedded
backend and 500 `DBCONVERTERROR` on Qdrant — the same client call, two answers.

### Deleting

`DELETE` is idempotent in both backends (`decisions/delete-is-idempotent.md`):
removing a fact that is not stored succeeds, so a retry does not fail because the
first attempt worked. `openapi.yaml` says the same — `removeFactByID` declares
`200` and `401` only, while `getFactByID` declares `404`. The one `400` a delete
can produce is ogen rejecting an empty `{factID}` before the handler, which is the
same for both backends and already covered above.

Consequences for the adapters:

- `kv` reads xvec's `ErrNotFound` — the batch error **and** the per-key one — as
  success; any other error is still returned;
- Qdrant issues the delete without an existence check; the endpoint accepts a point
  that is not there, so `pointExists` has no business in this operation;
- an ID that cannot name a fact returns success before the store is opened.

A `404` leaking out of `RemoveFactByID` is a bug: it breaks retries and
contradicts the spec. Presence reporting belongs to the read.

## Accepted differences — do not report, do not "fix"

These are deliberate. Each row is a decision; a row is only valid while its
decision record says the same thing.

| #   | Difference                                                                                                                                                                                | Why                                                                                                          | What must still hold                                                                                                                                                                                                 |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | ID shape: `kv` stores the full 64-hex SHA-256 digest; Qdrant stores the first 32 hex as a 36-character UUID. `GetFactByID` echoes the requested ID on `kv` and the stored UUID on Qdrant. | Qdrant accepts a point ID only as `u64` or UUID (`fact-id-from-text.md`)                                     | Both derive the ID from the text alone; both accept a full digest on lookup; recall returns whichever shape that backend stores; an ID that cannot name a fact here is absent on both — 404 on read, no-op on delete |
| 2   | Users: Qdrant keeps one collection per user (`memory-<userID>`, bare `memory` when empty); `kv` ignores `userID`.                                                                         | The embedded store is local files for one operator (`per-user-collections.md`)                               | Every method keeps the `userID` parameter; Qdrant never reads another user's collection, so a foreign fact is absent, not forbidden                                                                                  |
| 3   | Score direction: `kv` converts xvec's cosine distance with `1 - result.Score`; Qdrant's score is already a similarity and is used directly.                                               | xvec reports a distance, Qdrant a similarity (`semantic-recall-only.md`)                                     | `similarityThreshold` always means cosine similarity; never invert on Qdrant, never skip the inversion on `kv`                                                                                                       |
| 4   | Duplicate detection: `kv` rejects atomically inside one insert; Qdrant checks existence and then upserts, in two requests, and the loser sees success.                                    | Qdrant exposes no conditional insert over REST in the pinned version (`stored-records-immutable.md`)         | Duplicate answers `ErrAlreadyHasFact`; content and ID never change; the window is not "fixed" with a process-local lock                                                                                              |
| 5   | Provisioning: `kv` tries `Open` then `CreateAndOpen`; Qdrant does `GET /exists` then `PUT /collections`, tolerating `409`.                                                                | Both must create the store on first use, including on a read (`kv.getCollection`, `Qdrant.ensureCollection`) | A missing store never fails an operation; a concurrent create is not an error                                                                                                                                        |
| 6   | Failure surface: `kv` returns xvec errors; Qdrant returns `*httpError` with a status and the server's message.                                                                            | The contract has no transport error type                                                                     | Only the three sentinels are contractual; unmatched failures reach the client as 500 `INTERNAL_ERROR`; `ErrDBConvertError` never stands for caller input                                                             |
| 7   | Constructor and config shape differ (`kv.New(config, dataPath)` vs `qdrant.New(config, *http.Client)`); each declares its own `config`.                                                   | Only `scaffold` constructs them (`coupling-restrictions.md`)                                                 | No use case, controller or adapter names either package                                                                                                                                                              |
| 8   | Handle lifecycle: `kv` opens and closes a collection per operation; Qdrant reuses a shared `http.Client` and keeps no server handle.                                                      | A local index handle and an HTTP client are not the same resource                                            | No operation depends on state left by an earlier one; nothing leaks                                                                                                                                                  |

Anything not in this table is a divergence. Do not add a row to silence a finding:
a new accepted difference is a decision the user has to take, not a conclusion the
audit may reach on its own.

## Test parity

A test in one package and not the other means the untested backend is the one that
will drift. Every operation needs a test in both, driven by fakes — the testing
rules forbid network, filesystem and environment access:

- Qdrant: `fakeTransport` in `qdrant/qdrant_test.go` answers recorded requests.
- `kv`: `collectionOpener` (`kv/kv.go`) plus `fakeCollection` in
  `kv/getfact_test.go` replaces xvec.

Both suites must cover the same cases: duplicate, missing on read, **missing on
delete succeeding**, an unusable ID on read and on delete, wrong field type,
threshold and `topK`. Qdrant additionally covers per-user isolation, and `kv`
cannot — its storage has one user.

As of this writing `kv` covers `GetFactByID`, `RemoveFactByID` and the unusable-ID
guard; `AddFact` and `GetFactsList` have no tests and `kv/getfacts_test.go` is only
a TODO. Re-check before reporting; do not assume this is still true.

## Procedure

1. Read the seven decision records listed above.
2. Run Check 1: method counts, signatures, compile-time assertions, composition.
3. Run Check 2: the operation table, then the schema greps, then the sentinels.
4. Re-read the accepted table; confirm every row still describes the code. If a row
   no longer holds, that is a finding, not a licence to edit the row.
5. Check test parity.
6. Verify (below).
7. Report, separating the categories below.

## Reporting

- **Interface drift** — a signature that differs, a method only one backend has, a
  missing compile-time assertion, or a backend named outside `scaffold`.
- **Behaviour drift** — operation, input, `kv` outcome, Qdrant outcome, the outcome
  both should share, and the file:line on each side.
- **Schema drift** — key, metric, datatype, dimension or timestamp that differs.
- **Unusable IDs** — an input one backend answers as absent and the other accepts,
  fails on, or reports as a conversion error.
- **Open decisions** — a difference no decision record covers, and a row of the
  accepted table the code no longer matches. These are for the user to decide.
- **Accepted** — normally say nothing. Mention a row only if the code no longer
  matches it.

Quote `file:line` for both sides of a divergence. "The backends differ" is not a
finding; "`kv/getfact.go:31` returns 404 for `abc`, `qdrant/id.go:83` returns
`ErrDBConvertError`" is.

A difference a client sees has to be traced the whole way down: the HTTP
operation in `internal/controllers/web/methods.go` or the MCP tool in
`internal/adapters/servers/mcp/tools/`, then the use case, then the adapter. Say
which layer introduces it — a difference that exists only between the adapters is
the one this skill owns.

## Verify

```sh
gofmt -l internal/repositories
go build ./...
CGO_ENABLED=0 go test ./internal/repositories/... ./internal/use_cases/...
```

`make test` also rebuilds the UI because of `go:embed all:ui/dist`; use it when a
change reached outside `internal/repositories`.

A green suite proves nothing about parity: each adapter can pass its own tests
while disagreeing with the other. The comparisons above are the check.

## Recording a correction

Runs only when the **user** says a difference is wrong or intended.

- A difference the user accepts → a row in **Accepted differences**, with the
  reason and what still has to hold. Then, if the reason is not already in
  `decisions/`, write the decision record in the `## Decision` / `## Why` format
  from `AGENTS.md` and cite it in the row.
- An accepted difference the user revokes → delete the row and fix the code; do not
  leave a "was accepted" note.
- A rule about which IDs are absent → the `### IDs that cannot name a fact` section,
  and `decisions/fact-id-from-text.md` if the decision itself moved.

Rewrite the section instead of appending a story. Every run reads this whole file,
so it has to stay instructions, and git already keeps the history. Never record a
rule the user did not correct: a false rule is worse than none, because later runs
will follow it.
