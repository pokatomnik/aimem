---
name: architecture-audit
description: Audits aimem's layering rules — that every layer meets the others through narrow unexported interfaces, that only cmd/aimem/scaffold.go composes implementations, that no controller imports an adapter or the reverse, that no adapter imports another adapter, and that each layer keeps its own field spelling (CamelCase in Go, camelCase in the contract, snake_case in storage). Fixes violations it can fix and reports the rest. Use after adding or moving a package, after wiring a new dependency, when a review question is "does this respect the architecture", or when a build works but a layer seems to know about another.
---

# Architecture Audit

Enforce the one rule this repository cannot afford to drift on: **layers meet
through interfaces, never through implementations, and only `scaffold`
composes.**

A violation always compiles and always passes the tests — that is precisely why it
needs an audit instead of a review.

## The rule, stated once

A package depends on the behaviour it needs, declared as a narrow unexported
interface next to the consumer, and receives a concrete value injected from above.
It never names a concrete type from another layer — not in a signature, not in a
struct field, not in a `new(...)` call, not in a type assertion.

`cmd/aimem/scaffold.go` is the only composition root. It builds repositories,
adapters, use cases and controllers, and wires them.

### Allowed import directions

```
cmd/aimem/scaffold.go ──► everything (the composition root)

controllers/ ──► use_cases/ ──► entities/
     │                │              ▲
     └────────────────┴──────────────┘
                 (interfaces only)

repositories/ ──► entities/          (declares its own seams)
adapters/     ──► entities/, shared/ (declares its own seams)
```

Nothing may import `repositories/` or `adapters/` except `scaffold`. Exceptions
are about **data, not behaviour**:

| Importable everywhere  | Why                                                                      |
| ---------------------- | ------------------------------------------------------------------------ |
| `internal/entities/**` | `fact.FactFullView`, `user.User`, `apperrors` sentinels, `ollama` types. |
| `internal/shared/**`   | Cross-cutting helpers, by design.                                        |
| `cmd/aimem` `main`     | Entry point of the composition root.                                     |

A package may import its own subpackages (`servers/mcp` → `servers/mcp/tools`).
That is not a cross-layer edge.

## The audit

Run all six checks. Report every hit; a hit is a bug, not a style note.

### 1. A controller importing an adapter, or the reverse

```sh
grep -rn "internal/adapters" --include=*.go internal/controllers/ | grep -v ui/node_modules
grep -rn "internal/controllers" --include=*.go internal/adapters/
```

Both must be empty. The `web` controller receives the MCP endpoint as a plain
`http.Handler` field for exactly this reason; naming the MCP adapter's type there
would be the violation.

### 2. One adapter importing another, or one controller importing another

```sh
grep -rn "internal/adapters/" --include=*.go internal/adapters/ \
  | grep -v "adapters/servers/mcp/"
```

Any package pair that is not parent→child is a violation. `ollama` importing
`servers/mcp` to reuse a helper is a bug even when the helper is trivial; put the
shared code in `internal/entities` or `internal/shared`.

### 3. A use case importing a repository, adapter or controller

```sh
grep -rn "internal/repositories\|internal/adapters\|internal/controllers" \
  --include=*.go internal/use_cases/
```

Must be empty. A use case declares the interfaces it consumes (see
`internal/use_cases/recall`, which declares `ai` and its store) and `scaffold`
satisfies them.

### 4. Only the composition root imports implementations

```sh
grep -rln "internal/repositories\|internal/adapters" --include=*.go . \
  | grep -v "_test.go" | grep -v node_modules
```

Every result must be `cmd/aimem/scaffold.go` (or `cmd/aimem`'s own files), or a
file inside the very package being imported — plus `servers/mcp` importing its own
`tools` and `mcpdeps` subpackages. Anything else is a violation.

`_test.go` files are excluded from this grep and from the ones around it. Tests
are not held to the same hard rule: a test may reach for a concrete type where
that is genuinely simpler, and doing so welds nothing, because test code ships in
no binary. Prefer an interface seam anyway — a test that constructs a real
repository drags in the I/O the testing rules forbid — but treat a test-only
cross-layer import as a smell to weigh, not a defect to fix on sight.

### 5. No `WithAuth(bool)`-style flag in place of a decision

```sh
# a constructor that takes the mode as an argument — the decision leaks to callers
grep -rnE '\bNew(With|Without)?\w*\([^)]*\bauthRequired bool' --include=*.go internal/ \
  | grep -v "_test.go"
# a per-call authentication check inside a handler or tool
grep -rn "RequireAuthentication" --include=*.go internal/adapters internal/controllers \
  | grep -v "_test.go"
```

Both must be empty. `authRequired` **is** allowed to exist:

- as a field of `Controller` and `Methods`, set once by `NewWithAuth` or
  `NewWithoutAuth`;
- as a parameter of internal helpers inside the same package that pass the
  already-decided value down a call chain (`newSPAHandler`, `injectAuthFlag`).

What is not allowed is a constructor that accepts the mode as a `bool`, forcing
every caller to pass the same answer, or a handler/tool that asks
`RequireAuthentication` per call. `scaffold` takes the decision once and chooses
a constructor; nothing downstream re-decides.

### 6. Each layer uses its own field spelling

One value has three names, and a layer that uses the wrong one is a bug:

| Layer                                                         | Spelling     | Example      |
| ------------------------------------------------------------- | ------------ | ------------ |
| Go code                                                       | `CamelCase`  | `CreatedAt`  |
| `openapi.yaml`, generated server, UI client, MCP tool schemas | `camelCase`  | `createdAt`  |
| `kv` fields and Qdrant payload                                | `snake_case` | `created_at` |

The switch happens **between** layers — in `methods.go` when a domain value is
copied into a generated response, and in the repositories when a payload is read
— never within one layer.

```sh
# storage must stay snake_case
grep -rn 'keyField\w* *= *"[a-z]*[A-Z]' --include=*.go internal/repositories/
# the contract and the MCP schemas must stay camelCase
grep -rnE '\b(created_at|updated_at)\b' internal/controllers/web/openapi.yaml \
  internal/adapters/servers/mcp/tools/
```

Both must be empty. The Go field is `CreatedAt` whichever spelling carries it.

Both generated sides of the contract come from one `openapi.yaml`, so they are
only correct if regenerated together:

```sh
make api-generate   # ogen → server/  and  openapi-typescript → schema.d.ts
```

Running only one generator, or editing a generated file by hand, is the bug that
target exists to prevent. `tsc` catches it eventually, as a stale `schema.d.ts`.

## What to fix, and what to report

**Fix immediately:** an import that violates directions 1–4. There is exactly one
correct shape and no judgment is needed — move the type into an interface next to
its consumer and inject the value from `scaffold`.

**Report and wait:** anything that changes a contract or a decision — renaming a
wire field that already has stored data, collapsing two constructors into one,
adding a new shared interface to `entities` instead of a local one. The code is
written by hand and may be deliberate.

A cross-layer import inside a `_test.go` file is neither: report it as a
preference, with the interface seam that would remove it, and let the user
decide. The hard rule protects the binary; a test never reaches one.

When you fix a violation, the fix is not "make the import legal by loosening a
type". It is: declare the narrowest interface the consumer actually uses, next to
the consumer, and let `scaffold` pass the implementation.

## Verify

```sh
gofmt -l .
go build ./...
go vet ./...
CGO_ENABLED=0 go test ./...
```

`make test` also rebuilds the UI, because `go:embed all:ui/dist` does not compile
without `dist/`. If a UI field was renamed, run `npm run api:generate` first or
`tsc` will fail on the stale `schema.d.ts`.

A green build proves nothing about this audit — every violation in this file
compiles. The greps above are the check.

## Recording a correction

Runs only when the **user** says a decision was wrong. Put the result where the
next run will read it — never in a separate lessons file:

- a rule this skill must follow → the audit section it governs, one line, with
  the reason in a clause;
- a decision about the project → the matching file in `decisions/`, rewritten in
  the `## Decision` / `## Why` format described in `AGENTS.md`.

Rewrite the section instead of appending a story. Every run reads this whole file,
so it has to stay instructions, and git already keeps the history. Never record a
rule without its reason: an unexplained rule is the one the next run talks itself
back out of.
