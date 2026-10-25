# aimem

`aimem` is a small command-line long-term memory for AI assistants. It stores
facts locally, generates embeddings through [Ollama](https://ollama.com), and
retrieves the most relevant facts for a question using a local vector index
([xvec](https://github.com/gorse-io/xvec), HNSW with cosine similarity).

Everything runs on your machine: Ollama does the embedding and summarization,
and the vector database lives in `~/.aimem`. There are no external services and
no API keys.

## Features

- `remember` — save a fact; it is summarized and embedded.
- `recall` — find facts semantically similar to a question.
- `fact` — print a fact by its ID.
- `remove` — delete a fact by its ID.
- `import` — bulk-import facts from JSON Lines read from stdin.
- `mcp` — run the MCP server over `stdio` or `http` for use by AI assistants.
- `web` — serve the same memory over a small HTTP API.

## Requirements

- Go 1.27.1 or newer (see `go.mod`).
- [Ollama](https://ollama.com) running locally or reachable over HTTP.
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
`~/.config/aimem/config.toml`. The template lives in the repository at
`cmd/aimem/config/config.toml.example`:

```sh
mkdir -p ~/.config/aimem
cp cmd/aimem/config/config.toml.example ~/.config/aimem/config.toml
```

If the file is missing, `aimem` prints the template and the expected path to
stderr before it exits.

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

# Host the web API listens on
webHttpHost = "127.0.0.1:8888"
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
| `webHttpHost`         | Address for the `web` HTTP API. Defaults to `127.0.0.1:8888`.                                                                                |

> `recall` applies `topK` first and then filters by `similarityThreshold`, so a
> query can return fewer than `topK` results, or none at all, even when related
> facts exist.

> `dimensions` must match the number of values returned by the embedding model.
> If you change it after the database has been created, recreate `~/.aimem`.

## Usage

```sh
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

### Web API

`aimem web` exposes the same memory over HTTP, listening on `webHttpHost`
(`127.0.0.1:8888` by default):

```sh
aimem web
```

| Method | Path                    | Body                       | Returns                       |
| ------ | ----------------------- | -------------------------- | ----------------------------- |
| `POST` | `/api/v1/fact`          | `{"value":{"fact":"..."}}` | `{"summary":"..."}`           |
| `GET`  | `/api/v1/fact/{factID}` | —                          | `{"value":{...full fact...}}` |
| `POST` | `/api/v1/search`        | `{"query":"..."}`          | `{"value":{"found":[...]}}`   |

`search` runs the same semantic lookup as `recall`, so `topK` and
`similarityThreshold` apply and each item carries `id`, `createdAt`, and
`summary` — never the full text; use `GET /api/v1/fact/{factID}` for that.
Errors come back as `{"error":"<CODE>","message":"..."}`; the status is part of
the mapping: `400` for a malformed request, `404` for an unknown route or a
missing fact, `409` when the fact already exists, `422` when Ollama returns no
usable embedding, `502` when Ollama itself fails, and `500` for anything else.

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

All data is stored in `~/.aimem`:

- the xvec collection with embeddings and metadata;
- `plain.jsonl`, an append-only log of saved facts.

`plain.jsonl` records additions only: removing a fact deletes it from the
collection but leaves the line in the log. Delete the whole directory to start
from scratch.

## Development

```sh
make build    # build build/aimem for the current platform
make test     # run go test ./...
make generate # regenerate the ogen code in internal/controllers/web/server
make clean    # remove the build directory
```

Run `aimem` without arguments (or with `--help`) to list the available
commands.

## License

Released under the [MIT License](LICENSE).
