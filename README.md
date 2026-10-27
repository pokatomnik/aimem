# aimem

`aimem` is a small command-line long-term memory for AI assistants. It stores
facts, generates embeddings through [Ollama](https://ollama.com), and retrieves
the most relevant facts for a question by meaning, using cosine similarity.

Two storage backends are available and the config chooses between them:

- **Embedded** (default) — [xvec](https://github.com/gorse-io/xvec), an HNSW
  index kept in `~/.aimem`. No extra service, no API key.
- **Qdrant** — a Qdrant instance over HTTP, for running the memory outside one
  machine. Configure `[qdrant] host` to switch; see
  [Start the vector database](#3-start-the-vector-database-qdrant).

Ollama does the embedding and summarization in both cases. With the embedded
backend everything runs on your machine.

## Features

- `init` — write a default configuration file.
- `remember` — save a fact; it is summarized and embedded.
- `recall` — find facts semantically similar to a question.
- `fact` — print a fact by its ID.
- `remove` — delete a fact by its ID.
- `import` — bulk-import facts from JSON Lines read from stdin.
- `mcp` — run the MCP server over `stdio` or `http` for use by AI assistants.
- `web` — serve the same memory over an HTTP API and a browser UI.

## Requirements

- Go 1.27.1 or newer (see `go.mod`).
- [Node.js](https://nodejs.org) and `npm`, because the web UI is built from
  `internal/controllers/web/ui` and embedded into the binary with `go:embed`.
  `make build` and `make test` both run `npm ci` and `vite build` first, so a
  bare `go build ./...` on a fresh clone fails until `make ui-build` has run.
- [Ollama](https://ollama.com) running locally or reachable over HTTP.
- [Docker](https://docs.docker.com/get-docker/) with the Compose plugin, only if
  you want to use the bundled Qdrant service.
- An embedding model and a text-generation model pulled in Ollama, for example:

  ```sh
  ollama pull embeddinggemma-2:740m
  ollama pull gemma4:e2b
  ```

  Any model that supports the `embed` API can be used for embeddings. The
  summarization model can be any instruction-following model.

## Setup

### 1. Build and install

```sh
make build     # produces build/aimem
make install   # builds and copies the binary to GOBIN (or GOPATH/bin)
```

The build disables CGO (`CGO_ENABLED=0`) and strips debug information and build
metadata (`-trimpath -buildvcs=false -ldflags="-s -w -buildid="`) to keep the
binary as small as possible. The resulting binary is self-contained.

`make install` uses `$(go env GOBIN)` if it is set, otherwise
`$(go env GOPATH)/bin`. You can override the destination explicitly:

```sh
make install INSTALL_DIR=/usr/local/bin
```

Make sure the install directory is in your `PATH`.

### 2. Create the configuration

The configuration file is required and must be located at
`~/.config/aimem/config.toml`. There is no template file in the repository: the
binary generates the config from the built-in defaults, comments included.

```sh
aimem init                    # writes the file
```

If the file is missing, `aimem` writes it, prints where it landed, and exits —
no other command works until the file exists. `aimem init` does the same thing on
demand; if a file is already there it asks before overwriting.

Then adjust it:

```toml
# Ollama API endpoint
ollamaHost = "http://localhost:11434"

# Model used to create embeddings (required for remember/recall/import)
embeddingModel = "embeddinggemma-2:740m"

# Model used to summarize saved facts (required for remember/import)
summarizationModel = "gemma4:e2b"

# Number of facts returned by recall
# TopK is an upper bound; the similarity filter may return fewer, including none.
topK = 5

# Minimum cosine similarity for a fact to be returned by recall
# min: 0 (everything matches), max: 1 (exact match only)
similarityThreshold = 0.5

# Embedding dimensions; must match the embedding model output
dimensions = 768

# Host the MCP HTTP transport listens on
mcpHttpHost = "127.0.0.1:7777"

# Host the web API and UI listen on
webHttpHost = "127.0.0.1:8888"

# Optional. Setting `host` switches storage from the embedded index to Qdrant.
[qdrant]
host = ""          # e.g. "http://127.0.0.1:6333"
apiKey = ""        # leave empty when Qdrant has no authorization
```

| Key                   | Description                                                                                                                                  |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `ollamaHost`          | Base URL of the Ollama server.                                                                                                               |
| `embeddingModel`      | Model used through Ollama's `/api/embed` endpoint.                                                                                           |
| `summarizationModel`  | Model used through `/api/generate` to produce short summaries.                                                                               |
| `topK`                | Upper bound on the number of results returned by `recall`.                                                                                   |
| `similarityThreshold` | Minimum cosine similarity for a fact to be returned; `0` matches everything, `1` requires an exact match. Values outside `0..1` are clamped. |
| `dimensions`          | Vector size; it is fixed in the database schema when it is created.                                                                          |
| `mcpHttpHost`         | Address for the MCP HTTP transport. Defaults to `127.0.0.1:7777`.                                                                            |
| `webHttpHost`         | Address for the `web` HTTP API and UI. Defaults to `127.0.0.1:8888`.                                                                         |
| `[qdrant].host`       | Qdrant base URL. **Empty means the embedded backend**; any non-empty value switches to Qdrant.                                               |
| `[qdrant].apiKey`     | Sent as the `api-key` header. Leave empty when the server has no key — the header is then omitted entirely.                                  |

Unknown keys are ignored silently, so a typo leaves the field at its zero value
(`topK = 0` returns nothing, `similarityThreshold = 0` matches everything) with
no error. Compare your file against the table when something behaves oddly.

> `recall` applies `topK` first and then filters by `similarityThreshold`, so a
> query can return fewer than `topK` results, or none at all, even when related
> facts exist.

> `dimensions` must match the number of values returned by the embedding model.
> If you change it after the database has been created, recreate it — delete
> `~/.aimem` for the embedded backend, or drop the Qdrant collection.

### 3. Start the vector database (Qdrant)

The repository ships a `docker-compose.yaml` that runs [Qdrant](https://qdrant.tech)
with default settings, published on loopback only. All it needs is Docker:

```sh
cp .env.example .env      # optional: override ports, image tag, payload storage
docker compose up -d      # start in the background
docker compose down       # stop; data stays in the named volumes
docker compose down -v    # stop and delete the data
```

Check readiness before pointing anything at it:

```sh
curl http://127.0.0.1:6333/readyz    # all shards are ready
curl http://127.0.0.1:6333/healthz   # healthz check passed
```

The API is then at `http://127.0.0.1:6333` (REST) and `127.0.0.1:6334` (gRPC), and
the web UI at `http://127.0.0.1:6333/dashboard`.

#### Environment variables

Values come from `.env` (Compose picks it up from the directory holding
`docker-compose.yaml`) or from the environment. Every variable is optional and
falls back to the default in the table; leaving one empty in `.env` also gives the
default. `.env` is git-ignored, `.env.example` is the committed template with a
comment per variable.

| Variable                    | Default     | Description                                                                       |
| --------------------------- | ----------- | --------------------------------------------------------------------------------- |
| `QDRANT_VERSION`            | `v1.15.4`   | Image tag. Bump deliberately — major upgrades may need a data migration.          |
| `QDRANT_HTTP_PORT`          | `6333`      | Host port for the REST API.                                                       |
| `QDRANT_GRPC_PORT`          | `6334`      | Host port for the gRPC API.                                                       |
| `QDRANT_HTTP_HOST`          | `127.0.0.1` | Host address the REST port is bound to. `0.0.0.0` exposes it beyond this machine. |
| `QDRANT_GRPC_HOST`          | `127.0.0.1` | Same, for gRPC.                                                                   |
| `QDRANT_API_KEY`            | _(empty)_   | Key enabling write and read access. See the warning below before setting it.      |
| `QDRANT_READ_ONLY_API_KEY`  | _(empty)_   | Separate key limited to reads.                                                    |
| `QDRANT_ON_DISK_PAYLOAD`    | `true`      | Keep payload and its indexes on disk instead of in RAM.                           |
| `QDRANT_LOG_LEVEL`          | `INFO`      | `TRACE`, `DEBUG`, `INFO`, `WARN`, or `ERROR`.                                     |
| `QDRANT_TELEMETRY_DISABLED` | `false`     | `true` stops reporting usage statistics to the vendor.                            |

Inside the container the ports stay `6333`/`6334`; the variables only move the
host-side mapping. Data lives in the `qdrant_storage` and `qdrant_snapshots`
volumes, so `docker compose down` keeps it and `down -v` wipes it.

Qdrant has no authentication by default, which is why the ports are bound to
`127.0.0.1`. If you change `QDRANT_HTTP_HOST`/`QDRANT_GRPC_HOST` to `0.0.0.0`,
turn authentication on at the same time.

> `QDRANT_API_KEY` alone does nothing: Qdrant reads `service.api_key` as an
> optional string, and an _empty_ value counts as a key — it turns authentication
> on with a blank password and the app can no longer connect. So the two
> `QDRANT__SERVICE__API_KEY` lines in `docker-compose.yaml` are commented out
> rather than fed `${QDRANT_API_KEY:-}`. To enable a key: put it in `.env`,
> uncomment those two lines, and restart. `/healthz`, `/readyz`, `/livez` and
> `/dashboard` stay reachable even with a key set, so a readiness probe passing
> does not prove your client is authenticated.

No healthcheck is defined: the image is built on `debian-slim` without `curl` or
`wget`, and a port check would show the socket is open without proving the shards
are ready — use `/readyz` from the host instead.

#### Shape the collection must have

Storing aimem's memory in Qdrant requires a collection mirroring the current `kv`
module's schema, so the repository keeps its contract (`AddFact`, `GetFactByID`,
`GetFactsList`, `RemoveFactByID`):

| Item         | Value                                                                        |
| ------------ | ---------------------------------------------------------------------------- |
| Collection   | `memory`                                                                     |
| Distance     | `Cosine`                                                                     |
| Vector       | `embedding`, size `dimensions` from `config.toml` (default `768`), `float32` |
| Point ID     | derived from the fact's SHA-256 — Qdrant takes only a UUID or `u64`          |
| Payload keys | `summary`, `fact`, `created_at` (all strings)                                |

Two details differ from xvec and are handled by the adapter, not by this compose
file. The point ID format is one of them — see
[How a fact gets its ID](#how-a-fact-gets-its-id). The other is the score: Qdrant's
cosine score is already a similarity, so the `1 - result.Score` inversion that
`kv` applies must not be applied on top of it.

#### How a fact gets its ID

A fact's ID is derived from its text, never assigned separately. The text is
SHA-256 hashed, and everything else follows from that digest — which is what makes
the ID and the content unable to disagree.

The two backends need different ID formats, so the digest is shaped differently
for each:

| Backend  | Stored ID                             | Example prefix                         |
| -------- | ------------------------------------- | -------------------------------------- |
| Embedded | the full 64-character hex digest      | `4a44dc15364204a80fe80e9039455cc1...`  |
| Qdrant   | the first 32 hex characters as a UUID | `4a44dc15-3642-04a8-0fe8-0e9039455cc1` |

**Why the difference?** The embedded backend accepts any string as a primary key,
so it stores the digest verbatim. Qdrant accepts a point ID only as an unsigned
integer or a UUID, so the digest cannot be used as-is. Keeping the first 32 hex
characters — 128 bits of the digest — is enough to stay collision-free in
practice, and the ID remains deterministic: the same text always yields the same
point.

That truncation is safe because the UUID parser Qdrant uses does not validate the
version or variant bits. The result is not a well-formed v4 UUID — it looks like
one and is accepted as one, which is all Qdrant requires.

Consequences worth knowing:

- **IDs differ between backends.** The same fact stored in `~/.aimem` and in
  Qdrant has different IDs (`4a44dc…` vs `4a44dc15-…`). Both backends accept
  either form on lookup: a full digest is reduced to its UUID form, so data
  carried over from the embedded store resolves to the same point.
- **Prefer full digests when moving data.** A 64-character digest is unambiguous
  and each backend maps it to its own form. A short prefix is only meaningful in
  the backend that printed it.
- **The 128-bit prefix is the identity in Qdrant.** Two facts whose digests share
  the first 32 hex characters would collide. That needs a deliberate SHA-256
  prefix collision, which is not a practical concern for memory data.

#### Records are immutable

Immutability is a property of the stored data, not of one backend. It holds for
the embedded xvec index and for Qdrant alike, and any storage added later must
keep it.

A fact is written once and never patched in place. Writing the same text again is
answered with `db already has fact` instead of overwriting it. A fact that ever
had to change would be removed and inserted again — delete, then write — so a
stored record is always the result of exactly one write.

The rule follows from the identity: a fact's ID is the SHA-256 digest of its
text, so the ID and the content always agree. Rewriting the content while keeping
the ID would break that correspondence. Keeping records whole also removes a
whole class of problems: no half-updated payload, no vector and payload from
different writes, no question of which field the last writer touched.

For Qdrant this is enforced, not assumed. The repository touches only four
endpoints — insert, retrieve, query and delete — and `immutability_test.go` runs
every operation and fails if any request reaches a payload or vector update
endpoint (`/points/payload`, `/points/vectors`, `/points/batch`). Storing a
half-updated record is therefore a test failure, not a matter of discipline.

**Known limitation.** `AddFact` checks whether the point exists and then inserts
it, in two requests. Two concurrent writers of the same text can both pass the
check; the second insert replaces the record with the same ID and the same
content, only `created_at` is refreshed. The data stays correct — both writers
hold the same fact, so the record is still exactly one write's worth of content —
but the loser sees success instead of `db already has fact`.

Closing that window would need a conditional insert. Qdrant keeps one internally
but does not expose it over REST or gRPC in the pinned version, and its
`update_filter` guards updates of existing points rather than refusing an insert.
A process-local lock would not help: a second instance of `aimem` is a separate
process and would not share the lock, so it would only hide the window. Writes are
far rarer than reads, so the window is accepted rather than patched.

> Both backends are wired. `scaffold` builds the embedded repository and the
> Qdrant one, then selects Qdrant when `[qdrant] host` is non-empty. The compose
> file only stands up the server; pointing the config at it is what switches the
> storage.

## Usage

```sh
# Write a default configuration file (prompts before overwriting).
aimem init

# Save a fact. The summary is printed back.
aimem remember "The production database is hosted in Frankfurt"

# Find related facts, ordered by similarity.
aimem recall "Where is the production database?"

# Show a single fact by ID (IDs are printed by recall).
aimem fact 4a44dc15364204a80fe80e9039455cc1...

# Delete a fact by ID. This is irreversible.
aimem remove 4a44dc15364204a80fe80e9039455cc1...

# Import facts from JSON Lines (one JSON object per line).
aimem import < facts.jsonl
```

### MCP server

`aimem mcp` runs the MCP server for AI assistants. Two transports are available:

```sh
# stdio, for editors that spawn the server themselves
aimem mcp stdio

# HTTP, listening on mcpHttpHost (127.0.0.1:7777 by default)
aimem mcp http
```

The server exposes `remember`, `recall`, `read_fact`, and `remove_fact_by_id`. Its
usage instructions are embedded from
`internal/adapters/servers/mcp/INSTRUCTIONS.md` and sent to the client on
initialization.

### Web API and UI

`aimem web` serves the HTTP API and the browser UI from one port, listening on
`webHttpHost` (`127.0.0.1:8888` by default):

```sh
aimem web
```

Open `http://127.0.0.1:8888` for the UI — search, add, read and delete facts. It
is a single-page app using hash routing, embedded into the binary at build time.

The API contract lives in `internal/controllers/web/openapi.yaml`.

| Method   | Path                    | Body                       | Returns                       |
| -------- | ----------------------- | -------------------------- | ----------------------------- |
| `POST`   | `/api/v1/fact`          | `{"value":{"fact":"..."}}` | `{"value":{"summary":"..."}}` |
| `GET`    | `/api/v1/fact/{factID}` | —                          | `{"value":{...full fact...}}` |
| `DELETE` | `/api/v1/fact/{factID}` | —                          | `{"value":{}}`                |
| `POST`   | `/api/v1/search`        | `{"query":"..."}`          | `{"value":{"found":[...]}}`   |

Every success body is wrapped in `value`, including `POST /api/v1/fact`, whose
reply is `{"value":{"summary":"..."}}` — not a bare `{"summary":"..."}`.

`search` runs the same semantic lookup as `recall`, so `topK` and
`similarityThreshold` apply and each item carries `id`, `createdAt`, and
`summary` — never the full text; use `GET /api/v1/fact/{factID}` for that.

Errors come back as `{"error":"<CODE>","message":"..."}`; the status is part of
the mapping:

| Status | When                                                                      |
| ------ | ------------------------------------------------------------------------- |
| `400`  | Malformed request — bad JSON, a body failing validation, a short `query`. |
| `404`  | Unknown route, or a fact that does not exist (`GET` and `DELETE` alike).  |
| `409`  | The fact already exists (`ALREADY_HAS_FACT`).                             |
| `422`  | Ollama returned no usable embedding.                                      |
| `502`  | Ollama itself failed.                                                     |
| `500`  | Anything else, including a storage conversion error.                      |

Any path under `/api/` is answered as JSON, so a version typo like
`/api/v2/search` returns a JSON 404 instead of the UI's HTML.

### Import format

Each input line must be a JSON object with a `data` field:

```json
{"data":"The production database is hosted in Frankfurt"}
{"data":"Deploys run on every weekday at 10:00 UTC"}
```

Facts are identified by a SHA-256 hash of their content, so importing the same
fact twice fails with `db already has fact`. Empty lines are not skipped — make
sure the input contains valid JSON on every line.

## Data location

With the **embedded** backend all data is stored in `~/.aimem`:

- the xvec collection with embeddings and metadata;
- `plain.jsonl`, an append-only log of saved facts.

`plain.jsonl` records additions only: removing a fact deletes it from the
collection but leaves the line in the log. Delete the whole directory to start
from scratch.

With **Qdrant** the facts live in the collection described above and nothing is
written to `~/.aimem`. Note that `plain.jsonl` is a feature of the embedded
backend only — the Qdrant repository does not append to it. To reset, drop the
collection; `docker compose down -v` wipes the bundled one entirely.

Stored records are immutable on both backends — see
[Records are immutable](#records-are-immutable).

## Development

```sh
make build        # ui-build, then build/aimem for the current platform
make test         # ui-build, then go test ./...
make generate     # regenerate the ogen code in internal/controllers/web/server
make clean        # remove build/ and the UI dist

make ui-install   # npm ci (only when package.json or the lockfile changes)
make ui-build     # vite build into internal/controllers/web/ui/dist
make ui-test      # vitest run
make ui-lint      # eslint, zero warnings
make ui-typecheck # tsc --noEmit
```

The UI is embedded with `go:embed all:ui/dist`, so it must exist before the Go
compiler runs — that is why `build` and `test` depend on `ui-build`.

After changing `internal/controllers/web/openapi.yaml`, regenerate both sides:

```sh
make generate                       # ogen, for the Go server
cd internal/controllers/web/ui && npm run api:generate   # openapi-typescript
```

The generated `schema.d.ts` is committed, so a contract change that skips the
second step leaves the client typed against the old shapes.

Run `aimem` without arguments (or with `--help`) to list the available
commands.

## License

Released under the [MIT License](LICENSE).
