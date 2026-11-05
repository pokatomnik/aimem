# Repository Guidelines

## Project Structure & Module Organization

`aimem` is a Go CLI that stores facts, embeds them in-process with a bundled
EmbeddingGemma 2 model, and retrieves them by cosine similarity from one of two
storage backends: an embedded HNSW index (`xvec`) or Qdrant. Summaries are
produced by an external chat completions server. Module:
`github.com/pokatomnik/gomem` — note the module path does **not** match the
directory name.

Code is layered, with dependencies pointing inward:

- `cmd/aimem/` — entrypoint, `cobra` wiring (`scaffold.go`) and runtime config
  loading (`config.go`) from `AIMEM_*` process environment variables. Mandatory
  settings and defaults live in the `Config` env tags; `AIMEM_LLM_HOST`,
  `AIMEM_SUMMARIZATION_MODEL` and `AIMEM_DIMENSIONS` use `notEmpty`, so an empty
  value stops startup rather than failing at the first request. There is no
  `init` command or generated config template.
- `internal/controllers/cli/<command>/` — thin CLI layer; parses args and prints
  via `internal/shared/colors`. The only command is `web`. Remembering, searching,
  reading and removing facts live in the HTTP API and the SPA; there is no CLI
  command for them. There is no separate MCP command either: the MCP endpoint is
  mounted on the `web` server.
- `internal/controllers/web/` — the `web` command's HTTP API: `openapi.yaml` is
  the contract, `server/` is generated from it by `ogen` (`make generate`), and
  `methods.go`, `controller.go` plus `error_handler.go` hold the hand-written
  side. `handler.go` splits one listener three ways by path prefix; see
  `internal/controllers/web/README.md` for the diagram of that split.
- `internal/controllers/web/userid.go` — the `userIDKey` type that carries the
  authenticated user through `context.Context`. The key lives next to the code
  that reads it, not in a shared package.
- `internal/controllers/web/ui/` — the React SPA, built by Vite into `ui/dist` and
  embedded via `go:embed all:ui/dist` in `spa.go`. Its client types come from
  `openapi.yaml` through `npm run api:generate`.
- `internal/use_cases/<operation>/` — business logic, one package per operation
  (`recall`, `remember`, `getfact`, `removefact`); declares the narrow interfaces
  it consumes.
- `internal/entities/` — domain types (`fact`, `embeddings`, `user`) and sentinel
  errors (`apperrors`).
- `internal/repositories/users/` — parses `AIMEM_USERS_JSON` and resolves a
  bearer token to a user. It is used only for Qdrant authentication and is
  ignored by the embedded backend.
- `internal/repositories/kv/` — the embedded backend. The xvec collection sits
  behind the unexported `collection` interface so tests can fake it.
- `internal/repositories/qdrant/` — the same storage contract backed by Qdrant over
  HTTP; keeps the `kv` schema and satisfies the use cases' interfaces.
- `internal/adapters/` — outbound integrations (`embeder`: the in-process
  embedding model, `llm`: the chat completions client) and inbound input
  (`input`), plus the MCP server-side adapter (`servers/mcp`: the server, its
  tools, and the streamable-HTTP transport that exposes it).
- `internal/shared/` — cross-cutting helpers.

Runtime data lives in `~/.aimem` for the embedded backend. Runtime settings come
from `AIMEM_*` process environment variables; there is no config file, and the
Qdrant user list travels in `AIMEM_USERS_JSON` like every other setting.

## Architecture Overview

Three layers with dependencies pointing inward: `controllers` (Cobra, I/O) → `use_cases` (logic) → `entities` (domain types and sentinel errors). Repositories and adapters sit at the edge; use cases reach them only through unexported interfaces they declare, satisfied structurally. `scaffold` is the single composition root — a new command means two new packages and one wiring edit. Most controllers are stateless functions; the `web` controller is a struct built by `NewWithAuth(...)` or `NewWithoutAuth(...)` because it carries its dependencies, its config, and the authentication mode.

## Accepted Decisions

`decisions/` holds the decisions this project has already taken, one topic per
file, each with the rule and the reason behind it. The records are binding: a rule
written there outranks your instinct, and contradicting one is a decision to
revisit with the user, not an improvement to make on the way past.

Search that directory before changing how a layer meets another, how a field is
named in the contract or in storage, how a fact is identified, written or deleted,
how recall ranks results, how users are separated, how a request is routed or
authenticated, or how configuration selects a backend. List the directory and pick
the file by name — the names are the index, so there is nothing to read in full and
nothing to consult for unrelated work.

Write a new record when a choice is made that a later change would otherwise undo
or re-litigate:

- **One topic per file, named after the topic** — `coupling-restrictions.md`,
  `stored-records-immutable.md`. The name is how the next agent finds it, so it
  must say what the decision is about without opening the file.
- **English only.**
- **A short title, then two sections, in this order, and nothing else:**

  ```md
  # <The decision, as a noun phrase>

  ## Decision

  <the rule, as it stands now>

  ## Why

  <the reason, and the alternative that was rejected>
  ```

- **At most 100 lines.** A decision that needs more is more than one decision.
- **Rewrite, do not append.** When a decision changes, replace its text and let
  git keep the history. A superseded paragraph is exactly the stale information
  this directory exists to avoid.

Instructions that have to be followed on every task stay in `AGENTS.md` or in a
`SKILL.md`. A decision record explains a choice; it never replaces a rule.

## Prompts and Documentation

Two files are embedded and sent to a model; both must describe behavior that actually exists:

- `internal/adapters/servers/mcp/INSTRUCTIONS.md` — the MCP server's usage instructions, embedded in `servers/mcp/server.go`. Keep it aligned with the tools: whenever a change affects the server's purpose, tools, inputs, outputs, behavior, or usage triggers, update it in the same change. Keep it concise and focused on when and how the model should use the tools.
- `internal/adapters/llm/SUMMARIZATION_SYSTEM.md` — the summarization system prompt, embedded in `adapters/llm/consts.go` and sent by `llm.Summarize`.

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
- Interfaces are unexported and declared next to their consumer (`type embedder interface`).
- Import aliases follow `<layer><Name>Package` in `scaffold.go` (`aEmbedderPackage`, `aLLMPackage`, `ucRecallPackage`, `rKVPackage`). Match this pattern when adding commands.
- Constructors are `New(...)`; config uses `GetX()` getters satisfied by small local interfaces.
- `cobra` commands set `SilenceErrors`/`SilenceUsage`; return errors up to `main`, which prints `Error: ...` and exits 1.

## Testing Guidelines

Tests live beside the code as `<file>_test.go` and target the unexported interfaces.

**Every bug fix ships with a unit test.** A bug is behavior the code does not have yet, so the regression test is the only proof that it behaves that way now. Write the test first and watch it fail for the reason the bug describes, then change the code until it passes. A fix without a failing-first test is incomplete, and a test that already passed before the fix does not cover the bug.

Tests are hermetic. Assume the test runner has no permissions at all: no writable directory, no network, no configured environment. A test that needs any of that is a broken test, not a slow one. So no I/O from test code:

- **No environment** — no `os.Getenv`, `os.Setenv`, `os.Environ`, `t.Setenv`, and nothing that reads config from the environment. Configuration arrives as a fake satisfying the `GetX()` interface.
- **No network** — no live model server, no `http.Get`, no `net.Dial`, no `net.Listen`, and no `httptest.NewServer` (it binds a loopback socket). In-memory transports and `httptest.NewRecorder` are fine.
- **No filesystem** — no `t.TempDir()`, no `os.ReadFile`/`os.WriteFile`/`os.OpenFile`, no fixtures or golden files on disk. Use in-memory equivalents (`fstest.MapFS`, buffers, plain structs).

The single exception is `go:embed`: embedded data is resolved by the compiler and baked into the binary, so a test may read it — that is compiler work, not test logic, and it touches no disk at run time. `server_test.go` reading `INSTRUCTIONS.md` is the sanctioned example.

When a test would seem to require any of the above, **write the mock or stub first** — in the same change, before the test — and inject it. If the code under test reaches for the OS or a real backend directly, extract the seam first: declare a narrow unexported interface next to its consumer (the way `type embedder interface` does) and satisfy it with a fake.

Do not test library behavior. A test that drives xvec (or any dependency) to check that the dependency works is testing upstream code that is already tested upstream, and it drags in the I/O the rules above forbid. Assert on this project's logic only; for `GetFactsList` that means the `1 - result.Score` inversion, the threshold filter, `topK` as an upper bound, and the conversion errors — fed from a fake collection, not a real one.

When you fake a dependency, reproduce **every** channel the real call returns — values _and_ error. `xvec.Delete` of a missing key returns a non-empty result list _and_ a batch error wrapping `xvec.ErrNotFound`; a fake that models only one of them lets a leak through. A unit test that disagrees with an end-to-end run means the fake is wrong, not the live behavior. Fake errors that the code compares with `errors.Is` must wrap the same sentinel (`fmt.Errorf("...: %w", xvec.ErrNotFound)`), never a text-identical `errors.New`.

Run with `make test`. When you delete a phrase that a test asserts on, update the test in the same change.

## Architecture Audit Skill

`.agents/skills/architecture-audit/` holds a skill that verifies the layering
rules recorded in `decisions/coupling-restrictions.md` and
`decisions/field-naming-per-layer.md`: the allowed import directions, that only
`cmd/aimem/scaffold.go` imports implementations, and that wire names stay
`snake_case`. Run it after adding or moving a package, after wiring a new
dependency, or whenever a build works but one layer seems to know about another.
Every violation compiles, so a green build is not evidence — the greps in the
skill are the check.

## Commit & Pull Request Guidelines

History uses short imperative subjects, sometimes with leading bullets describing several changes (`- Change ID format`). Keep subjects under ~72 characters; add a body for multi-part work. Call out in the PR body any change touching storage, config keys, or command names, since those are user-visible and, for stored data, irreversible.
