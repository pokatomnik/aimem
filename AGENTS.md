# Repository Guidelines

## Project Structure & Module Organization

`aimem` is a Go CLI that stores facts locally, embeds them through Ollama, and retrieves them from an HNSW vector index (`xvec`). Module: `github.com/pokatomnik/gomem` — note the module path does **not** match the directory name.

Code is layered, with dependencies pointing inward:

- `cmd/aimem/` — entrypoint, `cobra` wiring (`scaffold.go`), config loading (`config.go`).
- `internal/controllers/<command>/` — thin CLI layer; parses args and prints via `internal/shared/colors`.
- `internal/use_cases/<command>/` — business logic; declares the narrow interfaces it consumes. Split by concern, e.g. `consume/{use_case,ai,unpack,fact}.go`.
- `internal/entities/` — domain types (`fact`, `ollama`) and sentinel errors (`apperrors`).
- `internal/repositories/kv/` — storage.
- `internal/adapters/` — outbound integrations (`ollama`) and inbound input (`input`).
- `internal/shared/` — cross-cutting helpers.

Runtime data lives in `~/.aimem`; config is required at `~/.config/aimem/config.toml` (`config.toml.example` is the template).

## Architecture Overview

Three layers with dependencies pointing inward: `controllers` (Cobra, I/O) → `use_cases` (logic) → `entities` (domain types and sentinel errors). Repositories and adapters sit at the edge; use cases reach them only through unexported interfaces they declare, satisfied structurally. `scaffold` is the single composition root — a new command means two new packages and one wiring edit. Controllers are stateless functions; use cases are structs built by `New(...)`.

## Architectural Constraints

- **Single fixed storage.** One local dataset, no namespaces or profiles. Config is file-only, with no flags or env overrides.
- **Recall is semantic only**, returning summaries — no tags, filters, or keyword search.
- **No cancellation or timeouts.** Requests run without a deadline, so a slow backend blocks the CLI.
- **Import is sequential**, and the write orchestration is duplicated between `remember` and `import`.

## MCP Server Instructions

`internal/adapters/servers/mcp/INSTRUCTIONS.md` is embedded and sent to the model as the MCP server's usage instructions. Keep it aligned with the MCP server and tool implementation: whenever a code change affects the server's purpose, available tools, inputs, outputs, behavior, or usage triggers, update this file in the same change. Keep the instructions concise and focused on when and how the model should use the tools, not on documenting implementation details.

## Build, Test, and Development Commands

```sh
make build    # CGO_ENABLED=0 build to build/aimem, stripped
make install  # build and copy to GOBIN (override: INSTALL_DIR=...)
make test     # CGO_ENABLED=0 go test ./...
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

There are no tests yet. Place them beside the code as `<file>_test.go` (e.g. `internal/use_cases/remember/use_case_test.go`), target the unexported interfaces, and prefer fakes over network calls to Ollama. Run with `make test`.

## Commit & Pull Request Guidelines

History uses short imperative subjects, sometimes with leading bullets describing several changes (`- Change ID format`). Keep subjects under ~72 characters; add a body for multi-part work. Call out in the PR body any change touching storage, config keys, or command names, since those are user-visible and, for stored data, irreversible.
