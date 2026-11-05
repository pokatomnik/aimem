# aimem

`aimem` is a small command-line long-term memory for AI assistants. It stores
facts, embeds them locally, and retrieves the most relevant facts for a question
by meaning, using cosine similarity.

Embeddings are computed in-process by EmbeddingGemma 2: the weights are
downloaded once into `~/.aimem/models/gemma` on first use and then reused
offline. No embedding service is involved. Short summaries come from any server
that speaks the chat completions API, addressed by `AIMEM_LLM_HOST`.

Two storage backends are available; `AIMEM_QDRANT_HOST` chooses between them:

- **Embedded** (default) — [xvec](https://github.com/gorse-io/xvec), an HNSW
  index kept in `~/.aimem`. No extra service, no API key.
- **Qdrant** — a Qdrant instance over HTTP, for running the memory outside one
  machine. Set `AIMEM_QDRANT_HOST` to switch; see
  [Start the vector database](#4-start-the-vector-database-qdrant).

With the embedded backend and a local model server, the only traffic that leaves
the machine is the one-time model download.

## Features

- `web` — serve the HTTP API, the browser UI, and the MCP endpoint from one port.
- Through the UI, the API or the MCP tools: save a fact (summarized and
  embedded), search by meaning, read one fact in full, delete a fact.
- With Qdrant, bearer-token authentication resolves each request to a user, and
  each user gets a separate collection. The embedded backend is anonymous.
- Two interchangeable storage backends: an embedded vector index, or Qdrant.

## Requirements

- Go 1.27.1 or newer (see `go.mod`).
- [Node.js](https://nodejs.org) and `npm`, because the web UI is built from
  `internal/controllers/web/ui` and embedded into the binary with `go:embed`.
  `make build` and `make test` both run `npm ci` and `vite build` first, so a
  bare `go build ./...` on a fresh clone fails until `make ui-build` has run.
- Network access for the first embedding: the model weights are fetched from
  HuggingFace into `~/.aimem/models/gemma` and reused offline afterwards. The
  model is fixed; there is no setting to choose another one.
- A server that speaks the chat completions API, reachable over HTTP — a local
  runtime or a hosted provider. This is the only network dependency at request
  time.
- [Docker](https://docs.docker.com/get-docker/) with the Compose plugin, only if
  you want to use the bundled Qdrant service.
- The name of a text-generation model on that server. It is sent as the `model`
  field, so it must be a name the server knows. Any instruction-following model
  works.

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

### 2. Set runtime configuration

`aimem` reads runtime settings from its process environment; it does not load
`.env` files itself. Set the required variables where the application runs:

```sh
export AIMEM_LLM_HOST="http://127.0.0.1:8000/v1"
export AIMEM_SUMMARIZATION_MODEL="your-model-name"
export AIMEM_DIMENSIONS="768"

# Optional: key for the model server
# export AIMEM_LLM_API_KEY="a-long-random-string"

# Optional: select Qdrant instead of the embedded backend
# export AIMEM_QDRANT_HOST="http://127.0.0.1:6333"

# Optional: Qdrant users and their tokens, as a JSON array
# export AIMEM_USERS_JSON='[{"id":"alice","tokens":["a-long-random-string"]}]'
```

Settings marked `Required` have no default and refuse an empty value; all others are optional:

| Variable                     | Required / default       | Description                                                                                                                                  |
| ---------------------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `AIMEM_LLM_HOST`             | Required, non-empty      | Base URL of the chat completions server, ending at the version segment, e.g. `http://127.0.0.1:8000/v1`. `/chat/completions` is appended.    |
| `AIMEM_LLM_API_KEY`          | Empty                    | Sent as `Authorization: Bearer` when set; the header is omitted when empty.                                                                  |
| `AIMEM_SUMMARIZATION_MODEL`  | Required, non-empty      | Model name sent in the `model` field when summarizing a fact.                                                                                |
| `AIMEM_TOP_K`                | `5`                      | Upper bound on the number of results returned by `recall`.                                                                                   |
| `AIMEM_SIMILARITY_THRESHOLD` | `0.7`                    | Minimum cosine similarity for a fact to be returned; `0` matches everything, `1` requires an exact match. Values outside `0..1` are clamped. |
| `AIMEM_DIMENSIONS`           | Required, non-empty      | Embedding vector size; it is fixed in the database schema when the collection is created. `768` for the bundled model.                       |
| `AIMEM_WEB_HTTP_HOST`        | `127.0.0.1:8888`         | Address for the `web` server: HTTP API, MCP endpoint, and UI.                                                                                |
| `AIMEM_QDRANT_HOST`          | Empty (embedded backend) | Qdrant base URL. Any non-empty value switches storage from the embedded index to Qdrant.                                                     |
| `AIMEM_QDRANT_API_KEY`       | Empty                    | Sent as the `api-key` header when set; omitted when empty.                                                                                   |
| `AIMEM_USERS_JSON`           | Empty                    | Qdrant users and their tokens as a JSON array; ignored by the embedded backend. See [Users and tokens](#3-add-users-and-tokens-qdrant-only). |

A typo in a variable name means that value is not applied. The three
`Required` settings must be present and non-empty, or startup fails;
`AIMEM_LLM_HOST`, `AIMEM_SUMMARIZATION_MODEL`, and `AIMEM_DIMENSIONS` are
checked that way, so a blank line in a deployment file stops the process with a
message naming the variable. Elsewhere an empty value falls back to the default.
`recall` applies `AIMEM_TOP_K` first and then filters by
`AIMEM_SIMILARITY_THRESHOLD`, so a query can return fewer results than the
limit, or none at all, even when related facts exist.

> `AIMEM_DIMENSIONS` must match the vector width of the embedding model — 768
> for the bundled one. If you change it after the database has been created,
> recreate it — delete `~/.aimem` for the embedded backend, or drop the Qdrant
> collections.

### 3. Add users and tokens (Qdrant only)

Whether the server requires a token is decided by the **storage backend**.
`AIMEM_USERS_JSON` supplies credentials for Qdrant; it does not enable auth for
the embedded backend:

| Backend         | `AIMEM_USERS_JSON` | Authentication                 |
| --------------- | ------------------ | ------------------------------ |
| Embedded (`kv`) | ignored            | not required — the API is open |
| Qdrant          | **required**       | every request needs a token    |

With the embedded backend there is nothing to partition: `~/.aimem` is one
operator's local files, and `userID` is not used by the storage at all. The
variable is simply ignored there, so nothing breaks if you set one and later
stay on the embedded backend.

With Qdrant each user gets a separate collection, so the server cannot answer a
request without knowing who is asking. **`aimem` refuses to start when Qdrant is
configured and the list defines no token**, with:

```
no users configured
```

That failure is deliberate: starting anyway would answer every request with
`401` and leave no way to fix it from the UI.

#### The format

`AIMEM_USERS_JSON` is a JSON array of users, each with an `id` and a list of
`tokens`:

```json
[
  { "id": "alice", "tokens": ["a-long-random-string"] },
  { "id": "bob", "tokens": ["another-random-string", "a-second-token-for-bob"] }
]
```

| Key      | Required | Description                                                            |
| -------- | -------- | ---------------------------------------------------------------------- |
| `id`     | yes      | Identifies the user. Becomes part of the Qdrant collection name.       |
| `tokens` | no       | Any number of tokens; each one authenticates the request as this user. |

`id` is the only field that describes the user; `tokens` are credentials, not
identity. A user with no tokens can be configured but will never authenticate.
The startup check counts tokens rather than entries, so a list in which no user
has a token is treated as no users at all and stops the server.

The value is parsed once, when the server starts, into a map from token to user.
Adding, changing or removing an entry takes effect on the next start.

#### The `id` format is the collection-name format

`id` is not free-form. It is used verbatim to build the collection name
`memory-<id>`, so it must satisfy Qdrant's collection naming rules:

- **ASCII letters, digits, `_` and `-` only.**
- No spaces, no dots, no `/`, no `@`, no non-ASCII characters.
- Practical bounds: keep it short, and do not start it with `_` or `-`.

```jsonc
// correct
{ "id": "alice" }
{ "id": "team-1_a" }

// wrong — Qdrant rejects the resulting collection name
{ "id": "alice@example.com" }
{ "id": "Иван" }
{ "id": "my team" }
```

A `/` is the worst case: it would address a different URL path entirely. An email
address is the most common mistake, since it looks like a natural identifier.
The adapter trims surrounding whitespace and nothing else, so what you write is
what reaches Qdrant.

Because each user owns a whole collection, two users storing the same text do not
collide — that is why the partition is per collection rather than a field in the
shared one.

A user can hold several tokens, which is how you rotate one: list the new token
next to the old one, restart, move the clients over, then drop the old token and
restart again. Every one of those steps needs a restart, because the list is
parsed once at startup into a token-to-user map rather than re-read per request.

### 4. Start the vector database (Qdrant)

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

| Item         | Value                                                               |
| ------------ | ------------------------------------------------------------------- |
| Collection   | `memory` for an empty user, `memory-<userID>` otherwise             |
| Distance     | `Cosine`                                                            |
| Vector       | `embedding`, size `AIMEM_DIMENSIONS` (required), `float32`          |
| Point ID     | derived from the fact's SHA-256 — Qdrant takes only a UUID or `u64` |
| Payload keys | `summary`, `fact`, `created_at` (all strings)                       |

Storage is partitioned per user: each userID gets its own collection with this
schema, created on first use. An empty userID keeps using the bare `memory`
collection, so a single-user setup and data written before users existed still
resolve. Because the point ID is derived from the fact's text alone, two users
storing the same text would collide in one shared collection — the per-user
collection is what keeps them apart, and it also means a fact belonging to
another user is simply absent rather than forbidden.

The embedded backend ignores the same userID: it is local files, single user.
The two backends share the public contract, not the storage layout.

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
- **An ID this backend cannot store is "not found".** A pasted or hand-edited
  string that is not a UUID or a digest cannot name a fact, so `GET` answers it
  with `404` and `FACT_DOES_NOT_EXIST`, exactly like an ID nobody stored — never
  with a `500`. The embedded backend applies the same rule to keys its index
  rejects. `DELETE` is a no-op for the same input, like any other absent fact.
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
> Qdrant one, then selects Qdrant when `AIMEM_QDRANT_HOST` is non-empty. The
> compose file only stands up the server; setting this environment variable is
> what switches the storage.

## Usage

The only CLI command is `web`, which serves the API, the UI and the MCP endpoint
from one port:

```sh
aimem web
```

Remembering, searching, reading and removing facts are done through the UI or
the API, not through the CLI. An MCP client gets the same four operations as
tools.

With Qdrant, every API call must carry `Authorization: Bearer <token>`; a request
without a valid token is answered with `401` and the appError body `{"error":
"UNAUTHORIZED"}`. The embedded backend is anonymous and does not require a token.

```sh
# Save a fact. For Qdrant, include the bearer token; embedded mode is anonymous.
# The summary comes back in the response.
curl -X POST http://127.0.0.1:8888/api/v1/fact \
  -H 'Authorization: Bearer a-long-random-string' \
  -H 'Content-Type: application/json' \
  -d '{"value":{"fact":"The production database is hosted in Frankfurt"}}'

# Search by meaning; returns summaries with their IDs.
curl -X POST http://127.0.0.1:8888/api/v1/search \
  -H 'Authorization: Bearer a-long-random-string' \
  -H 'Content-Type: application/json' \
  -d '{"query":"Where is the production database?"}'

# Read one fact in full by an ID from a search result.
curl http://127.0.0.1:8888/api/v1/fact/4a44dc15364204a80fe80e9039455cc1... \
  -H 'Authorization: Bearer a-long-random-string'
```

### MCP server

The MCP server is part of `aimem web`: it is mounted on the same listener, under
`/mcp`, next to the API and the UI. There is no separate command and no separate
port.

```sh
aimem web
```

Point an MCP client at `http://127.0.0.1:8888/mcp` using the streamable HTTP
transport. The path is matched by prefix, so anything beginning with `/mcp`
reaches the MCP handler rather than the UI.

The server exposes `remember`, `recall`, `read_fact`, and `remove_fact_by_id`. Its
usage instructions are embedded from
`internal/adapters/servers/mcp/INSTRUCTIONS.md` and sent to the client on
initialization.

The MCP tools authenticate the same way as the REST API. With Qdrant, every
`tools/call` must carry an `Authorization: Bearer <token>` header, and the token
decides which user's facts the call reads or writes. Tokens come from
`AIMEM_USERS_JSON`, described in
[Users and tokens](#3-add-users-and-tokens-qdrant-only); a missing or unknown
token is rejected without touching data. With the embedded backend, calls are
anonymous and do not require a token.

`AIMEM_WEB_HTTP_HOST` defaults to `127.0.0.1:8888`, and the MCP library rejects a
non-loopback `Host` header on a loopback listener; binding `AIMEM_WEB_HTTP_HOST`
to a non-loopback address publishes the MCP server to the network, so make sure
the tokens are meaningful before you do.

### Web API and UI

`aimem web` serves the HTTP API, the MCP endpoint and the browser UI from one
port, listening on `AIMEM_WEB_HTTP_HOST` (`127.0.0.1:8888` by default):

```sh
aimem web
```

Open `http://127.0.0.1:8888` for the UI — search, add, read and delete facts. It
is a single-page app using hash routing, embedded into the binary at build time.

#### The token prompt

The UI asks for a token only when the server requires one. There is no endpoint
for that answer: the server writes it into the page it serves, as a global the
entry script reads before it renders.

```html
<script>
  window.requireAuth = true;
</script>
```

The value comes from the same startup decision that gates the API, so the page and
the handlers cannot disagree:

- **Embedded backend** → `false`. The UI opens straight into the app.
- **Qdrant** → `true`. The UI shows a single "Access token" field and nothing else
  until one is saved; every request is then signed with it.

Only `index.html` is rewritten, under `no-cache`; hashed assets are served
byte-for-byte and keep their immutable caching. A page served without the global
(an older build, a static host) is read as `false` — the safe default, since
locking the user out of an API that needs no token is worse than one rejected
request.

The token is stored in the browser's `localStorage` under `aimem.token`, so it
survives a reload. A token is not checked when it is entered — the form cannot
know whether it is valid — so a wrong one surfaces as a failed request. When
that failure is `UNAUTHORIZED`, the error block offers **Logout** next to
**Retry**: it drops the stored token and shows the form again. Logging out is
purely a client-side operation, and it clears the query cache, since every entry
in it was fetched under the identity being dropped.

There is no session expiry, and no server-side logout: to revoke a token itself,
remove it from `AIMEM_USERS_JSON` and restart the server. The list is parsed once
at startup, so until that restart the token keeps working; afterwards any copy
left in a browser is a harmless stale string.

The API contract lives in `internal/controllers/web/openapi.yaml`.

| Method   | Path                    | Body                       | Returns                       |
| -------- | ----------------------- | -------------------------- | ----------------------------- |
| `POST`   | `/api/v1/fact`          | `{"value":{"fact":"..."}}` | `{"value":{"summary":"..."}}` |
| `GET`    | `/api/v1/fact/{factID}` | —                          | `{"value":{...full fact...}}` |
| `DELETE` | `/api/v1/fact/{factID}` | —                          | `{"value":{}}`                |
| `POST`   | `/api/v1/search`        | `{"query":"..."}`          | `{"value":{"found":[...]}}`   |

Every success body is wrapped in `value`, including `POST /api/v1/fact`, whose
reply is `{"value":{"summary":"..."}}` — not a bare `{"summary":"..."}`.

`search` runs the same semantic lookup as `recall`, so `AIMEM_TOP_K` and
`AIMEM_SIMILARITY_THRESHOLD` apply and each item carries `id`, `created_at`, and
`summary` — never the full text; use `GET /api/v1/fact/{factID}` for that.

Errors come back as `{"error":"<CODE>","message":"..."}`; the status is part of
the mapping:

| Status | When                                                                                                                         |
| ------ | ---------------------------------------------------------------------------------------------------------------------------- |
| `400`  | Malformed request — bad JSON, a body failing validation, a short `query`, or a missing `{factID}` segment (`/api/v1/fact/`). |
| `401`  | Qdrant is selected and no valid token came with the request.                                                                 |
| `404`  | Unknown route, or a fact that does not exist — an unnameable ID and an unstored one alike (`GET`; `DELETE` is a no-op).      |
| `409`  | The fact already exists (`ALREADY_HAS_FACT`).                                                                                |
| `422`  | The embedding step produced no vector (`EMPTY_EMBEDDINGS`).                                                                  |
| `502`  | The summarization server answered with an error or with no choices.                                                          |
| `500`  | Anything else, including a storage conversion error and a failure to compute an embedding.                                   |

`401` only happens with Qdrant. With the embedded backend the same endpoints
answer anonymously, because there is one local operator and nothing to
partition — a token sent there is ignored rather than rejected.

Any path under `/api/` is answered as JSON, so a version typo like
`/api/v2/search` returns a JSON 404 instead of the UI's HTML. The three
prefixes are matched in this order: `/mcp`, then `/api/`, then everything else,
which is the UI.

Authentication is declared once for the whole API: every operation inherits the
`bearerAuth` requirement from the document root, so `401` applies to all four
routes. It is checked before the body is decoded, which is why a request with no
token never reaches validation and cannot distinguish "bad JSON" from "no
credentials" — it simply gets `401`.

### Bulk insertion

There is no import command and no bulk endpoint. Insert facts one request at a
time. For Qdrant, set `TOKEN` to one of the configured user tokens; embedded mode
is anonymous:

```sh
jq -r '.data' facts.jsonl | while IFS= read -r line; do
  curl -s -X POST http://127.0.0.1:8888/api/v1/fact \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d "$(jq -Rn --arg f "$line" '{value:{fact:$f}}')"
done
```

Facts are identified by a SHA-256 hash of their content, so inserting the same
text twice fails with `db already has fact` — a `409` on the API. Retrying a
batch is therefore safe: the duplicates are rejected and the rest go through.

## Data location

With the **embedded** backend all data is stored in `~/.aimem`: the xvec
collection with embeddings and metadata, plus the embedding model under
`~/.aimem/models/gemma`. Delete the whole directory to start from scratch — the
model is downloaded again on the next embedding.

With **Qdrant** the facts live in the collections described above and nothing is
written to `~/.aimem` except that model directory, which both backends share.
To reset the facts, drop the collections; `docker compose down -v` wipes the
bundled instance entirely.

There is no export file. Facts are readable through the API, the UI and the MCP
tools; a bulk export was removed and is not part of the current behaviour.

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

After changing `internal/controllers/web/openapi.yaml`, regenerate both sides in
one step:

```sh
make api-generate   # ogen → server/  and  openapi-typescript → schema.d.ts
```

The generated `schema.d.ts` is committed, so a contract change that skips the
client half leaves the client typed against the old shapes.

Run `aimem` without arguments (or with `--help`) to list the available
commands.

## Architecture

`aimem` is layered, and the layers point inward:

```
cmd/aimem/scaffold.go          composition root — the only place that
        │                      builds implementations and wires them
        ▼
controllers/  ──►  use_cases/  ──►  entities/
(Cobra, HTTP,      (logic)          (domain types, sentinel errors)
 SPA)
        ▲
        │  injected through interfaces
repositories/ (kv, qdrant, users)      adapters/ (embeder, llm, input, mcp)
```

### Layers meet through interfaces, never through implementations

**This is the single most important rule in the codebase, and breaking it is a
bug that must be fixed immediately.**

A package depends on the behaviour it needs, declared as a narrow unexported
interface next to the consumer, and receives a concrete value injected from
above. It never names a concrete type from another layer — not in a signature,
not in a struct field, not in a `new(...)` call, not in a type assertion.

**Only `cmd/aimem/scaffold.go` may import implementations and compose them.** It
is the composition root: it builds the concrete repositories, adapters, use
cases and controllers, and wires them together. Nothing else may do that work.

This rule is easy to break by accident, because the shortcut always compiles and
always passes the tests:

| Not allowed                                                                           | Why it is a bug                                                      |
| ------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `controllers` importing `adapters`, or the reverse                                    | Welds a transport layer to an integration layer.                     |
| One adapter importing another, or one controller importing another                    | Each should depend on an interface in `entities` or declare its own. |
| A use case importing a repository package                                             | The use case declares what it needs; `scaffold` supplies it.         |
| Returning or accepting a concrete type from another layer where an interface would do | The caller now knows how the other layer is built.                   |
| A `WithAuth(bool)`-style flag threaded through callers                                | The decision belongs in `scaffold`, expressed as two constructors.   |

Narrow exceptions, all about data rather than behaviour: shared value types in
`internal/entities/**` and helpers in `internal/shared/**` are importable
everywhere, and `main` in `cmd/aimem` may import anything because it is the
composition root's entry point.

When in doubt, ask which side owns the behaviour: if "both", the interface
belongs in `entities` or `shared`; if "the caller", declare it next to the caller
and inject it.

### Field names

One value has three spellings, and each layer uses its own:

| Layer                                                                        | Spelling     | Example      |
| ---------------------------------------------------------------------------- | ------------ | ------------ |
| Go code                                                                      | `CamelCase`  | `CreatedAt`  |
| API contract — `openapi.yaml`, the generated server and UI client, MCP tools | `camelCase`  | `createdAt`  |
| Storage — the `kv` fields and the Qdrant payload                             | `snake_case` | `created_at` |

The JSON a client sees uses `camelCase`, like any other JSON API; the storage
schema keeps `snake_case`, so data already written to `~/.aimem` and to Qdrant
stays readable. A field therefore changes spelling _between_ layers — in
`methods.go` when the domain value is copied into the generated response, and in
the repositories when a payload is read — but never within one layer.

Both generated sides come from the same `openapi.yaml`, so they must be
regenerated together:

```sh
make api-generate   # ogen → server/  and  openapi-typescript → schema.d.ts
```

Editing a generated file by hand, or running only one of the two generators,
leaves the client typed against shapes the server no longer sends.

### One listener, three kinds of route

`aimem web` serves the API, the MCP endpoint and the UI from one port. Splitting
the path is `internal/controllers/web/handler.go`; see
[`internal/controllers/web/README.md`](internal/controllers/web/README.md) for the
diagram and the rules.

## License

Released under the [MIT License](LICENSE).
