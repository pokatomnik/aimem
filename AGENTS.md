# Repository Guidelines

## Project Structure & Module Organization

`aimem` is a Go CLI that stores facts locally, embeds them through Ollama, and retrieves them from an HNSW vector index (`xvec`). Module: `github.com/pokatomnik/gomem` — note the module path does **not** match the directory name.

Code is layered, with dependencies pointing inward:

- `cmd/aimem/` — entrypoint, `cobra` wiring (`scaffold.go`) and the embedded `config.toml.example` (`config/example.go`); config loading lives in `cmd/aimem/config/` (`config.go`, `init.go`).
- `internal/controllers/cli/<command>/` — thin CLI layer; parses args and prints via `internal/shared/colors`. Commands: `remember`, `recall`, `fact`, `remove`, `import`, `mcp {stdio,http}`, `web`.
- `internal/controllers/web/` — the `web` command's HTTP API: `openapi.yaml` is the contract, `server/` is generated from it by `ogen` (`make generate`), and `methods.go` plus `error_handler.go` hold the hand-written side.
- `internal/use_cases/<command>/` — business logic; declares the narrow interfaces it consumes. Split by concern, e.g. `consume/{use_case,ai,unpack,fact}.go`.
- `internal/entities/` — domain types (`fact`, `ollama`) and sentinel errors (`apperrors`).
- `internal/repositories/kv/` — storage; also writes `plain.jsonl`, an append-only log of additions.
- `internal/adapters/` — outbound integrations (`ollama`) and inbound input (`input`), plus the servers (`servers/{mcp,stdio,http}`).
- `internal/shared/` — cross-cutting helpers.

Runtime data lives in `~/.aimem`; config is required at `~/.config/aimem/config.toml` (`config.toml.example` is the template).

## Architecture Overview

Three layers with dependencies pointing inward: `controllers` (Cobra, I/O) → `use_cases` (logic) → `entities` (domain types and sentinel errors). Repositories and adapters sit at the edge; use cases reach them only through unexported interfaces they declare, satisfied structurally. `scaffold` is the single composition root — a new command means two new packages and one wiring edit. Most controllers are stateless functions; the `web` controller is a struct built by `New(...)` because it carries its dependencies and config.

## Architectural Constraints

- **Single fixed storage.** One local dataset, no namespaces or profiles. Config is file-only, with no flags or env overrides. Unknown config keys are ignored silently — a typo leaves the field at its zero value, so keep `config.toml.example` in sync with `Config`.
- **Recall is semantic only**, returning summaries — no tags, filters, or keyword search. `topK` is applied first, then results are filtered by `similarityThreshold`; xvec reports cosine as a distance (0 = identical), so the code converts it to a similarity before comparing.
- **No timeouts.** Requests run without a deadline, so a slow backend blocks the CLI. Cancellation exists only through signals: `main` wraps `context.Background()` in `signal.NotifyContext` (SIGINT/SIGTERM/SIGQUIT) and prints `Cancelled` when the context is canceled.
- **Import is sequential**, and the write orchestration is duplicated between `remember` and `import`.
- **`plain.jsonl` is append-only.** Removing a fact does not rewrite it, by design.

## Prompts and Documentation

Two files are embedded and sent to a model; both must describe behavior that actually exists:

- `internal/adapters/servers/mcp/INSTRUCTIONS.md` — the MCP server's usage instructions, embedded in `servers/mcp/server.go`. Keep it aligned with the tools: whenever a change affects the server's purpose, tools, inputs, outputs, behavior, or usage triggers, update it in the same change. Keep it concise and focused on when and how the model should use the tools.
- `internal/adapters/ollama/SUMMARIZATION_SYSTEM.md` — the summarization system prompt, embedded in `adapters/ollama/summarization.go`.

Tool names and descriptions live in `servers/mcp/tools/*.go`; `server_test.go` asserts that names match `INSTRUCTIONS.md` and that every tool is described. New prompts may appear elsewhere — check for `go:embed` and embedded instruction text before assuming this list is complete.

## Build, Test, and Development Commands

```sh
make build    # CGO_ENABLED=0 build to build/aimem, stripped
make install  # build and copy to GOBIN (override: INSTALL_DIR=...)
make test     # CGO_ENABLED=0 go test ./...
make generate # regenerate the ogen code in internal/controllers/web/server
make clean    # remove build/
```

## Coding Style & Naming Conventions

- Format with `gofmt`; no linter config exists, so `go vet ./...` is the practical check.
- Group type declarations in a single `type ( ... )` block.
- Interfaces are unexported and declared next to their consumer (`type ai interface`), keeping packages mockable.
- Import aliases follow `<layer><Name>Package` in `scaffold.go` (`aOllamaPackage`, `ucRecallPackage`, `rKVPackage`). Match this pattern when adding commands.
- Constructors are `New(...)`; config uses `GetX()` getters satisfied by small local interfaces.
- `cobra` commands set `SilenceErrors`/`SilenceUsage`; return errors up to `main`, which prints `Error: ...` and exits 1.

## Testing Guidelines

Tests live beside the code as `<file>_test.go` and target the unexported interfaces.

Tests are hermetic. Assume the test runner has no permissions at all: no writable directory, no network, no configured environment. A test that needs any of that is a broken test, not a slow one. So no I/O from test code:

- **No environment** — no `os.Getenv`, `os.Setenv`, `os.Environ`, `t.Setenv`, and nothing that reads config from the environment. Configuration arrives as a fake satisfying the `GetX()` interface.
- **No network** — no live Ollama, no `http.Get`, no `net.Dial`, no `net.Listen`, and no `httptest.NewServer` (it binds a loopback socket). In-memory transports and `httptest.NewRecorder` are fine.
- **No filesystem** — no `t.TempDir()`, no `os.ReadFile`/`os.WriteFile`/`os.OpenFile`, no fixtures or golden files on disk. Use in-memory equivalents (`fstest.MapFS`, buffers, plain structs).

The single exception is `go:embed`: embedded data is resolved by the compiler and baked into the binary, so a test may read it — that is compiler work, not test logic, and it touches no disk at run time. `server_test.go` reading `INSTRUCTIONS.md` is the sanctioned example.

When a test would seem to require any of the above, **write the mock or stub first** — in the same change, before the test — and inject it. If the code under test reaches for the OS or a real backend directly, extract the seam first: declare a narrow unexported interface next to its consumer (the way `type ai interface` does) and satisfy it with a fake.

Do not test library behavior. A test that drives xvec (or any dependency) to check that the dependency works is testing upstream code that is already tested upstream, and it drags in the I/O the rules above forbid. Assert on this project's logic only; for `GetFactsList` that means the `1 - result.Score` inversion, the threshold filter, `topK` as an upper bound, and the conversion errors — fed from a fake collection, not a real one.

Current coverage:

- `internal/adapters/servers/mcp/server_test.go` — tool names, descriptions, and the key guidance in `INSTRUCTIONS.md`, driven through an in-memory MCP transport. Reads the embedded text only; complies with the rules above.
- `internal/repositories/kv/getfacts_test.go` — **no tests**: the file is a `TODO` explaining that `GetFactsList` is uncovered and must be tested through a faked collection rather than a real xvec one. See the note above.

Run with `make test`. When you delete a phrase that a test asserts on, update the test in the same change.

## Commit & Pull Request Guidelines

History uses short imperative subjects, sometimes with leading bullets describing several changes (`- Change ID format`). Keep subjects under ~72 characters; add a body for multi-part work. Call out in the PR body any change touching storage, config keys, or command names, since those are user-visible and, for stored data, irreversible.
