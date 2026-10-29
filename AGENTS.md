# Repository Guidelines

## Project Structure & Module Organization

`aimem` is a Go CLI that stores facts, embeds them through Ollama, and retrieves
them by cosine similarity from one of two storage backends: an embedded HNSW
index (`xvec`) or Qdrant. Module: `github.com/pokatomnik/gomem` — note the module
path does **not** match the directory name.

Code is layered, with dependencies pointing inward:

- `cmd/aimem/` — entrypoint, `cobra` wiring (`scaffold.go`) and config loading
  (`config/`: `config.go`, `init.go`, `qdrant_config.go`). There is no
  `config.toml.example`: the template is produced by `DefaultConfig().String()`,
  which marshals the struct and its `comment` tags.
- `internal/controllers/cli/<command>/` — thin CLI layer; parses args and prints
  via `internal/shared/colors`. Commands: `init`, `web`. Remembering, searching,
  reading and removing facts live in the HTTP API and the SPA; there is no CLI
  command for them. There is no separate MCP command either: the MCP endpoint is
  mounted on the `web` server.
- `internal/controllers/web/` — the `web` command's HTTP API: `openapi.yaml` is
  the contract, `server/` is generated from it by `ogen` (`make generate`), and
  `methods.go`, `controller.go` plus `error_handler.go` hold the hand-written
  side. `handler.go` splits one listener three ways by path prefix.
- `internal/controllers/web/userid.go` — the `userIDKey` type that carries the
  authenticated user through `context.Context`. The key lives next to the code
  that reads it, not in a shared package.
- `internal/controllers/web/ui/` — the React SPA, built by Vite into `ui/dist` and
  embedded via `go:embed all:ui/dist` in `spa.go`. Its client types come from
  `openapi.yaml` through `npm run api:generate`.
- `internal/use_cases/<command>/` — business logic; declares the narrow
  interfaces it consumes.
- `internal/entities/` — domain types (`fact`, `ollama`, `user`) and sentinel
  errors (`apperrors`).
- `internal/repositories/users/` — reads the `[[users]]` entries of
  `config.toml` and resolves a bearer token to a user. Read from disk on every
  call, so an edit takes effect without a restart.
- `internal/repositories/kv/` — the embedded backend; also writes `plain.jsonl`, an
  append-only log of additions. The xvec collection sits behind the unexported
  `collection` interface so tests can fake it.
- `internal/repositories/qdrant/` — the same storage contract backed by Qdrant over
  HTTP; keeps the `kv` schema and satisfies the use cases' interfaces.
- `internal/adapters/` — outbound integrations (`ollama`) and inbound input
  (`input`), plus the MCP server-side integrations (`servers/mcp` for the server
  and its tools, `servers/http` for the streamable HTTP handler).
- `internal/shared/` — cross-cutting helpers.

Runtime data lives in `~/.aimem` for the embedded backend; config is required at
`~/.config/aimem/config.toml`.

## Architecture Overview

Three layers with dependencies pointing inward: `controllers` (Cobra, I/O) → `use_cases` (logic) → `entities` (domain types and sentinel errors). Repositories and adapters sit at the edge; use cases reach them only through unexported interfaces they declare, satisfied structurally. `scaffold` is the single composition root — a new command means two new packages and one wiring edit. Most controllers are stateless functions; the `web` controller is a struct built by `New(...)` because it carries its dependencies and config.

## Architectural Constraints

- **Storage backend is chosen by config.** Two backends implement the same
  contract: the embedded `xvec` index and Qdrant over HTTP. `scaffold` picks
  Qdrant when `[qdrant] host` is non-empty and the embedded one otherwise. There
  are no namespaces or profiles. Config is file-only, with no flags or env
  overrides. Unknown config keys are ignored silently — a typo leaves the field at
  its zero value, so keep the `comment` tags on `Config` accurate; they _are_ the
  template users copy.
- **A fact's ID is derived from its text**, so the ID and the content cannot
  disagree. The digest is shaped per backend: `kv` stores the full 64-character
  SHA-256 hex, while Qdrant takes the first 32 hex characters as a UUID (Qdrant
  accepts only a UUID or `u64`; its parser does not check the version or variant
  bits). Both accept a full digest on lookup, so `kv` data resolves to the same
  Qdrant point.
- **Stored records are immutable.** This holds for every backend, `kv` and
  `qdrant` alike, and for any storage added later: a fact is written once and never
  patched in place. Writing the same text again answers `ErrAlreadyHasFact` instead of overwriting. A change, if one is ever needed, is two steps — delete the record, then write the new one — so a stored record is always the result of exactly one write. Concretely: do not call in-place update APIs (payload update, vector update, partial or batch overwrite). `internal/repositories/qdrant/immutability_test.go` enforces this for the Qdrant backend by failing if any operation reaches a mutation endpoint.
  - The write path is read-then-write: `AddFact` checks whether the point exists and then inserts it, in two requests. Two concurrent writers of the same text can both pass the check; the loser then rewrites the same ID with the same content and sees success instead of `ErrAlreadyHasFact`. The data stays correct, so the window is accepted — writes are far rarer than reads. Do not paper over it with a process-local lock: a second instance is a separate process and would not share it, so the lock would only hide the window.
- **Recall is semantic only**, returning summaries — no tags, filters, or keyword search. `topK` is applied first, then results are filtered by `similarityThreshold`. The two backends report that number differently and the adapter must respect it: xvec returns a cosine _distance_ (0 = identical), so `kv` converts it with `1 - Score`, while Qdrant's cosine score is already a similarity and must not be inverted.
- **Users are separated by collection in Qdrant, not by a payload field.** All four operations take a `userID` and act on that user's collection, named by `collectionName` (`memory-<userID>`, or the bare `memory` for an empty one). The choice follows from the ID rule above: the point ID comes from the text alone, so in one shared collection two users storing the same text would collide and the second `AddFact` would answer `ErrAlreadyHasFact`. Per-user collections also make a foreign fact simply absent — `GetFactByID` returns `ErrFactDoesNotExist` rather than a permission error, with no ownership check needed. Do not move the partition into the payload: it would need a filter on every read and a matching index, and the retrieve endpoint accepts no filter at all. `kv` takes the same `userID` and ignores it on purpose — local files, one operator; the backends share the public contract, not the storage layout. Because a collection name becomes part of the URL path, a `userID` must be an opaque identifier: Qdrant allows only ASCII letters, digits, `_` and `-`, and a `/` would address a different route.
- **One listener serves three things.** `newHandler` in
  `internal/controllers/web/handler.go` splits requests by path prefix, in this
  order: `/mcp` → the MCP streamable-HTTP handler, `/api/` → the generated API,
  everything else → the SPA. The split lives in a handler rather than in
  `server.WithNotFound` on purpose: a typo inside the prefix, such as
  `/api/v2/search`, then still reaches the API and is answered with a JSON 404
  instead of quietly returning `index.html`. MCP is created once in `scaffold`
  and passed into the `web` controller, so the MCP endpoint shares the `web`
  port and authenticates every API request and every MCP tool call with a bearer
  token.
- **No timeouts.** Requests run without a deadline, so a slow backend blocks the CLI. Cancellation exists only through signals: `main` wraps `context.Background()` in `signal.NotifyContext` (SIGINT/SIGTERM/SIGQUIT) and prints `Cancelled` when the context is canceled.
- **Writes are one fact per request.** The API has no bulk endpoint, and the
  former `import` command no longer exists; a client that needs bulk insertion
  issues one `POST /api/v1/fact` per fact.
- **`plain.jsonl` is append-only.** Removing a fact does not rewrite it, by design.

## Prompts and Documentation

Two files are embedded and sent to a model; both must describe behavior that actually exists:

- `internal/adapters/servers/mcp/INSTRUCTIONS.md` — the MCP server's usage instructions, embedded in `servers/mcp/server.go`. Keep it aligned with the tools: whenever a change affects the server's purpose, tools, inputs, outputs, behavior, or usage triggers, update it in the same change. Keep it concise and focused on when and how the model should use the tools.
- `internal/adapters/ollama/SUMMARIZATION_SYSTEM.md` — the summarization system prompt, embedded in `adapters/ollama/summarization.go`.

Tool names and descriptions live in `servers/mcp/tools/*.go`; `server_test.go` asserts that names match `INSTRUCTIONS.md` and that every tool is described. New prompts may appear elsewhere — check for `go:embed` and embedded instruction text before assuming this list is complete.

## Build, Test, and Development Commands

```sh
make build    # ui-build, then CGO_ENABLED=0 build to build/aimem, stripped
make install  # build and copy to GOBIN (override: INSTALL_DIR=...)
make test     # ui-build, then CGO_ENABLED=0 go test ./...
make generate # regenerate the ogen code in internal/controllers/web/server
make clean    # remove build/ and the UI dist
```

The UI targets are `ui-install`, `ui-build`, `ui-test`, `ui-lint`,
`ui-typecheck` (each runs npm in `internal/controllers/web/ui`). `build` and
`test` depend on `ui-build` because `go:embed all:ui/dist` does not compile until
`dist/` exists.

## Coding Style & Naming Conventions

- Format with `gofmt`; no linter config exists, so `go vet ./...` is the practical check.
- Group type declarations in a single `type ( ... )` block.
- Interfaces are unexported and declared next to their consumer (`type ai interface`), keeping packages mockable.
- Import aliases follow `<layer><Name>Package` in `scaffold.go` (`aOllamaPackage`, `ucRecallPackage`, `rKVPackage`). Match this pattern when adding commands.
- Constructors are `New(...)`; config uses `GetX()` getters satisfied by small local interfaces.
- `cobra` commands set `SilenceErrors`/`SilenceUsage`; return errors up to `main`, which prints `Error: ...` and exits 1.

## Testing Guidelines

Tests live beside the code as `<file>_test.go` and target the unexported interfaces.

**Every bug fix ships with a unit test.** A bug is behavior the code does not have yet, so the regression test is the only proof that it behaves that way now. Write the test first and watch it fail for the reason the bug describes, then change the code until it passes. A fix without a failing-first test is incomplete, and a test that already passed before the fix does not cover the bug.

Tests are hermetic. Assume the test runner has no permissions at all: no writable directory, no network, no configured environment. A test that needs any of that is a broken test, not a slow one. So no I/O from test code:

- **No environment** — no `os.Getenv`, `os.Setenv`, `os.Environ`, `t.Setenv`, and nothing that reads config from the environment. Configuration arrives as a fake satisfying the `GetX()` interface.
- **No network** — no live Ollama, no `http.Get`, no `net.Dial`, no `net.Listen`, and no `httptest.NewServer` (it binds a loopback socket). In-memory transports and `httptest.NewRecorder` are fine.
- **No filesystem** — no `t.TempDir()`, no `os.ReadFile`/`os.WriteFile`/`os.OpenFile`, no fixtures or golden files on disk. Use in-memory equivalents (`fstest.MapFS`, buffers, plain structs).

The single exception is `go:embed`: embedded data is resolved by the compiler and baked into the binary, so a test may read it — that is compiler work, not test logic, and it touches no disk at run time. `server_test.go` reading `INSTRUCTIONS.md` is the sanctioned example.

When a test would seem to require any of the above, **write the mock or stub first** — in the same change, before the test — and inject it. If the code under test reaches for the OS or a real backend directly, extract the seam first: declare a narrow unexported interface next to its consumer (the way `type ai interface` does) and satisfy it with a fake.

Do not test library behavior. A test that drives xvec (or any dependency) to check that the dependency works is testing upstream code that is already tested upstream, and it drags in the I/O the rules above forbid. Assert on this project's logic only; for `GetFactsList` that means the `1 - result.Score` inversion, the threshold filter, `topK` as an upper bound, and the conversion errors — fed from a fake collection, not a real one.

When you fake a dependency, reproduce **every** channel the real call returns — values _and_ error. `xvec.Delete` of a missing key returns a non-empty result list _and_ a batch error wrapping `xvec.ErrNotFound`; a fake that models only one of them lets a leak through. A unit test that disagrees with an end-to-end run means the fake is wrong, not the live behavior. Fake errors that the code compares with `errors.Is` must wrap the same sentinel (`fmt.Errorf("...: %w", xvec.ErrNotFound)`), never a text-identical `errors.New`.

Current coverage:

- `internal/adapters/servers/mcp/server_test.go` — tool names, descriptions, and the key guidance in `INSTRUCTIONS.md`, driven through an in-memory MCP transport. Reads the embedded text only; complies with the rules above.
- `internal/repositories/kv/getfact_test.go` — the `collection` seam satisfies both `xvec.Collection` and a fake. Covers a missing fact (`ErrFactDoesNotExist`, never a nil dereference or a leaked xvec error), a present fact, a wrong field type (`ErrDBConvertError`), and that a real delete failure is not disguised as a missing fact. Fake errors wrap `xvec.ErrNotFound`, and the fake reproduces Delete's _both_ channels: a batch error and a per-key error.
- `internal/repositories/kv/getfacts_test.go` — **no tests**: the file is a `TODO` explaining that `GetFactsList` is uncovered and must be tested through the same faked collection rather than a real xvec one.
- `internal/repositories/qdrant/` — the four operations against a fake `http.RoundTripper`: request shapes (named vector, `wait=true`, `topK` as the limit), that the score is compared as a similarity without the kv inversion, duplicate and missing-fact answers, collection creation, and API-key propagation. `immutability_test.go` runs every operation and fails if any request reaches an in-place mutation endpoint.
- `internal/controllers/web/`, `internal/adapters/input/`, and `internal/controllers/web/ui/src/**/*.test.ts(x)` — the HTTP error mapping, the SPA handler, stdin parsing, and the UI (component, hook, and API-client tests).

Run with `make test`. When you delete a phrase that a test asserts on, update the test in the same change.

## Commit & Pull Request Guidelines

History uses short imperative subjects, sometimes with leading bullets describing several changes (`- Change ID format`). Keep subjects under ~72 characters; add a body for multi-part work. Call out in the PR body any change touching storage, config keys, or command names, since those are user-visible and, for stored data, irreversible.
