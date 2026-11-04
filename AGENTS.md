# Repository Guidelines

## Project Structure & Module Organization

`aimem` is a Go CLI that stores facts, embeds them through Ollama, and retrieves
them by cosine similarity from one of two storage backends: an embedded HNSW
index (`xvec`) or Qdrant. Module: `github.com/pokatomnik/gomem` — note the module
path does **not** match the directory name.

Code is layered, with dependencies pointing inward:

- `cmd/aimem/` — entrypoint, `cobra` wiring (`scaffold.go`) and runtime config
  loading (`config.go`) from `AIMEM_*` process environment variables. Required
  settings and defaults live in the `Config` env tags; there is no `init`
  command or generated config template.
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
- `internal/use_cases/<command>/` — business logic; declares the narrow
  interfaces it consumes.
- `internal/entities/` — domain types (`fact`, `ollama`, `user`) and sentinel
  errors (`apperrors`).
- `internal/repositories/users/` — parses `AIMEM_USERS_JSON` and resolves a
  bearer token to a user. It is used only for Qdrant authentication and is
  ignored by the embedded backend. The list is parsed once, on first use, and
  cached for the lifetime of the process, so a change takes effect on the next
  start.
- `internal/repositories/kv/` — the embedded backend. The xvec collection sits
  behind the unexported `collection` interface so tests can fake it.
- `internal/repositories/qdrant/` — the same storage contract backed by Qdrant over
  HTTP; keeps the `kv` schema and satisfies the use cases' interfaces.
- `internal/adapters/` — outbound integrations (`ollama`) and inbound input
  (`input`), plus the MCP server-side adapter (`servers/mcp`: the server, its
  tools, and the streamable-HTTP transport that exposes it).
- `internal/shared/` — cross-cutting helpers.

Runtime data lives in `~/.aimem` for the embedded backend. Runtime settings come
from `AIMEM_*` process environment variables; there is no config file, and the
Qdrant user list travels in `AIMEM_USERS_JSON` like every other setting.

## Architecture Overview

Three layers with dependencies pointing inward: `controllers` (Cobra, I/O) → `use_cases` (logic) → `entities` (domain types and sentinel errors). Repositories and adapters sit at the edge; use cases reach them only through unexported interfaces they declare, satisfied structurally. `scaffold` is the single composition root — a new command means two new packages and one wiring edit. Most controllers are stateless functions; the `web` controller is a struct built by `New(...)` because it carries its dependencies and config.

## Architectural Constraints

- **Layers meet through interfaces, never through implementations. This is the
  single most important rule in this repository.** A package depends on the
  behaviour it needs, declared as a narrow unexported interface next to the
  consumer (`type ai interface`, `type collection interface`), and gets a
  concrete value injected from above. It never names a concrete type from
  another layer — not in a signature, not in a struct field, not in a
  `new(...)` call, not in a type assertion.

  **Only `cmd/aimem/scaffold.go` may import implementations and compose them.**
  It is the composition root: it builds the concrete repositories, adapters, use
  cases and controllers and wires them together. Nothing else may do that work,
  and nothing else may take its place.

  This matters because the tempting shortcut is invisible in review: a
  controller importing a repository "just for the type", an adapter importing a
  controller "just for a constructor", or two adapters importing each other to
  reuse a helper all compile, pass their tests, and silently weld two layers
  together. An agent is more likely to make this mistake than a human, because
  the shortcut is always the shortest path to a working build.

  Concretely, these are bugs that must be fixed at once:

  - `internal/controllers/**` importing `internal/adapters/**`, or the reverse.
  - One adapter importing another (`ollama` → `servers/mcp`), or one controller
    importing another, instead of both depending on an interface in
    `internal/entities` or on a locally declared one.
  - A use case importing a repository package. Use cases declare the interfaces
    they need (`internal/use_cases/recall` declares `ai` and its store);
    `scaffold` satisfies them.
  - A package returning a concrete type from another layer, or accepting one as
    an argument, where an interface would do.
  - A `WithAuth(bool)`-style parameter threaded to avoid the composition root
    making a decision. The decision belongs in `scaffold`, expressed as two
    constructors, not as a flag every caller has to get right.

  Legitimate exceptions are narrow and always about data, not behaviour:
  shared value types in `internal/entities/**` (`fact.FactFullView`,
  `user.User`, the `apperrors` sentinels) may be imported everywhere, and
  `internal/shared/**` helpers are importable everywhere by design. `main` in
  `cmd/aimem` may import anything, because it is the composition root's entry.

  When in doubt, ask which side owns the behaviour. If the answer is "both", the
  interface belongs in `internal/entities` or `internal/shared`; if it is "the
  caller", declare it next to the caller and inject.

- **Two names for one field, and each layer uses its own.** The same value has a
  Go name, an external name and a storage name, and they are deliberately
  different conventions:

  | Layer                                                                                       | Spelling     | Example                     |
  | ------------------------------------------------------------------------------------------- | ------------ | --------------------------- |
  | Go code                                                                                     | `CamelCase`  | `CreatedAt`, `FactFullView` |
  | External contract — `openapi.yaml`, the generated Go server and UI client, MCP tool schemas | `camelCase`  | `createdAt`                 |
  | Storage — the `kv` fields and the Qdrant payload                                            | `snake_case` | `created_at`                |

  The external contract is what a client sees, so it follows the convention every
  JSON API uses; the storage schema is its own thing and keeps `snake_case`, so
  data already in `~/.aimem` and in Qdrant stays readable. A field therefore
  changes spelling **between** `methods.go` (which reads `CreatedAt` from the
  domain type and fills the generated one) and the repositories (which read
  `created_at` out of a payload) — never within one layer.

  The two generated sides of the contract come from one `openapi.yaml`, so they
  can only agree if regenerated together: `make api-generate` runs both ogen and
  `openapi-typescript`. Editing either generated tree by hand, or running only
  one generator, is the bug this target exists to prevent.

- **Storage backend is chosen by config.** Two backends implement the same
  contract: the embedded `xvec` index and Qdrant over HTTP. `scaffold` picks
  Qdrant when `AIMEM_QDRANT_HOST` is non-empty and the embedded one otherwise.
  Runtime settings are read from process environment variables by
  `cmd/aimem/config.go`; there are no CLI flags or `.env` loader. Qdrant users
  and their tokens come from `AIMEM_USERS_JSON`, a JSON array of
  `{"id": ..., "tokens": [...]}` entries parsed once at startup.
  `AIMEM_EMBEDDING_MODEL`, `AIMEM_SUMMARIZATION_MODEL`, and `AIMEM_DIMENSIONS`
  are required; defaults are defined by the env tags.
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
  and passed into the `web` controller as an `http.Handler`, so the MCP endpoint
  shares the `web` port and authenticates every API request and every MCP tool
  call with a bearer token.
- **No timeouts.** Requests run without a deadline, so a slow backend blocks the CLI. Cancellation exists only through signals: `main` wraps `context.Background()` in `signal.NotifyContext` (SIGINT/SIGTERM/SIGQUIT) and prints `Cancelled` when the context is canceled.
- **Writes are one fact per request.** The API has no bulk endpoint, and the
  former `import` command no longer exists; a client that needs bulk insertion
  issues one `POST /api/v1/fact` per fact.
- **Authentication follows the storage backend, not the config contents.**
  `users.RequireAuthentication` makes that one decision: with the embedded backend
  it returns `false` and every request runs anonymously (`AIMEM_USERS_JSON`, if
  set, is ignored, because `kv` ignores `userID`); with Qdrant it returns `true`
  and **fails to start** with `ErrNoUsersConfigured` when the list defines no
  token.

  It is asked once, in `scaffold`, with the context `main` provides — and the
  answer is expressed the same way for both servers: **two constructors each**, no
  bool argument, no per-call check.

  - `cmdWebPackage.NewWithAuth` / `NewWithoutAuth` set the unexported
    `authRequired` field.
  - `aMCPPackage.NewWithAuth` / `NewWithoutAuth` inject either
    `tools.NewTokenAuthenticator(users)` or `tools.NewAnonymousAuthenticator()`.

  Each MCP constructor returns an `*mcp.Server` that is a plain `http.Handler`:
  the streamable-HTTP transport is the adapter's own business, so the SDK's
  `*mcp.Server` never reaches `scaffold`. There is one MCP package, not one for
  the server and one for the transport — a second package would only be a package
  that builds the first, and it would misdescribe itself as "http" while the real
  HTTP server lives in `internal/controllers/web`.

  A tool calls `a.UserID(ctx, req)` and nothing else: which authenticator was
  injected _is_ the mode, so a tool holds no branch of its own and cannot disagree
  with the web API. Do not reintroduce a per-call `RequireAuthentication` in the
  tools, and do not add a `WithAuth(bool)` variant — a bool at the end of a long
  argument list is exactly the kind of thing that gets flipped by accident.

  The one branch lives in `scaffold` and picks both constructors from the same
  value, so the API and the tools cannot end up in different modes. Credential
  parsing (`tools.NewTokenAuthenticator`) stays apart from the decision: an
  authenticator turns a header into a user ID, it does not judge a deployment.

  The answer then travels as `Methods.authRequired` (which `Methods.getUserID`
  enforces by refusing an empty user) and as the `authRequired` baked into the
  served page. `getUserID` checks the mode _before_ reading the context: with
  auth off the user is never looked up, so a stale value cannot travel into a
  store that ignores it. Do not swap those two steps — reading the context first
  makes the function fetch something it has decided it will not use.

  Do not move the decision into the handlers: it has to exist before the first
  request, since a server that starts and then answers 401 forever gives the
  operator no way out.

- **The token prompt is baked into the served HTML, not exposed as an endpoint.**
  `newSPAHandler` takes `authRequired` and `injectAuthFlag` writes
  `<script>window.requireAuth=…</script>` into `index.html`'s head, before the
  entry module runs. Only the entry document is rewritten; hashed assets are
  served untouched so their immutable caching stays honest. The answer comes from
  the same `authRequired` value the handlers enforce, so the page and the API
  cannot disagree. Do not add a `GET /api/v1/auth` endpoint for this: the server
  already knows the answer when it builds the page, and an endpoint would let the
  two drift. The token lives in `localStorage` under `aimem.token`. Logging out
  is client-side only — the error block drops the token and shows the form again
  — so revoking a token for real means removing it from `AIMEM_USERS_JSON` and
  restarting, because the list is parsed once and cached.

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
- `internal/adapters/servers/mcp/handler_test.go` — the same adapter driven through a fake `http.RoundTripper` that invokes the handler in process, so the streamable-HTTP wiring is exercised without binding a socket: a token in the header reaches the tool as a user ID, a missing header is refused before any store is touched, and the anonymous wiring accepts the same call.
- `internal/repositories/kv/getfact_test.go` — the `collection` seam satisfies both `xvec.Collection` and a fake. Covers a missing fact (`ErrFactDoesNotExist`, never a nil dereference or a leaked xvec error), a present fact, a wrong field type (`ErrDBConvertError`), and that a real delete failure is not disguised as a missing fact. Fake errors wrap `xvec.ErrNotFound`, and the fake reproduces Delete's _both_ channels: a batch error and a per-key error.
- `internal/repositories/kv/getfacts_test.go` — **no tests**: the file is a `TODO` explaining that `GetFactsList` is uncovered and must be tested through the same faked collection rather than a real xvec one.
- `internal/repositories/qdrant/` — the four operations against a fake `http.RoundTripper`: request shapes (named vector, `wait=true`, `topK` as the limit), that the score is compared as a similarity without the kv inversion, duplicate and missing-fact answers, collection creation, and API-key propagation. `immutability_test.go` runs every operation and fails if any request reaches an in-place mutation endpoint.
- `internal/controllers/web/`, `internal/adapters/input/`, and `internal/controllers/web/ui/src/**/*.test.ts(x)` — the HTTP error mapping, the SPA handler, stdin parsing, and the UI (component, hook, and API-client tests).

Run with `make test`. When you delete a phrase that a test asserts on, update the test in the same change.

## Architecture Audit Skill

`.agents/skills/architecture-audit/` holds a skill that verifies the layering
rules above: the allowed import directions, that only `cmd/aimem/scaffold.go`
imports implementations, and that wire names stay `snake_case`. Run it after
adding or moving a package, after wiring a new dependency, or whenever a build
works but one layer seems to know about another. Every violation compiles, so a
green build is not evidence — the greps in the skill are the check.

## Commit & Pull Request Guidelines

History uses short imperative subjects, sometimes with leading bullets describing several changes (`- Change ID format`). Keep subjects under ~72 characters; add a body for multi-part work. Call out in the PR body any change touching storage, config keys, or command names, since those are user-visible and, for stored data, irreversible.
