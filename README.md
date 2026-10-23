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
- `import` — bulk-import facts from JSON Lines read from stdin.

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
`~/.config/aimem/config.toml`:

```sh
mkdir -p ~/.config/aimem
cp config.toml.example ~/.config/aimem/config.toml
```

Then adjust it:

```toml
# Ollama API endpoint
ollamaHost = "http://localhost:11434"

# Model used to create embeddings (required for remember/recall/import)
embeddingModel = "embeddinggemma-2:740m"

# Model used to summarize saved facts (required for remember/import)
summarizationModel = "gemma4:e2b"

# Number of facts returned by recall
topK = 5

# Embedding dimensions; must match the embedding model output
dimensions = 768
```

| Key                  | Description                                                         |
| -------------------- | ------------------------------------------------------------------- |
| `ollamaHost`         | Base URL of the Ollama server.                                      |
| `embeddingModel`     | Model used through Ollama's `/api/embed` endpoint.                  |
| `summarizationModel` | Model used through `/api/generate` to produce short summaries.      |
| `topK`               | Maximum number of results returned by `recall`.                     |
| `dimensions`         | Vector size; it is fixed in the database schema when it is created. |

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


# Import facts from JSON Lines (one JSON object per line).
aimem import < facts.jsonl
```

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
- `plain.jsonl`, a readable mirror of saved facts.

Delete this directory to start from scratch.

## Development

```sh
make build   # build build/aimem for the current platform
make test    # run go test ./...
make clean   # remove the build directory
```

Run `aimem` without arguments (or with `--help`) to list the available
commands.

## License

Released under the [MIT License](LICENSE).
