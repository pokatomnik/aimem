# ID Validity Is Backend-Local

## Decision

The contract declares `factID` as an unformatted `string`, and it stays that way.
Neither repository validates an ID against the contract; each asks only whether
the string can name a fact in _its_ storage:

- `kv` accepts any valid UTF-8 primary key of 1..`xvec.MaxPrimaryKeyBytes`
  (`validFactID`);
- `qdrant` accepts a canonical UUID, 32 hex characters, or a full 64-character
  digest, which it reduces to a UUID (`pointIDFromExternal`).

Three cases are settled, and a read answers each one deliberately:

| ID                           | `GET /api/v1/fact/{factID}` | Where the answer comes from |
| ---------------------------- | --------------------------- | --------------------------- |
| Empty (`/api/v1/fact/`)      | `400 BAD_REQUEST`           | ogen's parameter decoder    |
| Not nameable by this backend | `404 FACT_DOES_NOT_EXIST`   | the adapter                 |
| Nameable, but not stored     | `404 FACT_DOES_NOT_EXIST`   | the adapter                 |

`DELETE` answers `200` in all three, because deleting is idempotent — see
`delete-is-idempotent.md`. `401` takes precedence over all of it when the
deployment requires a token.

The adapter check runs before the store is opened, so an ID that cannot name a
fact provisions no collection and sends no request. It is never
`ErrDBConvertError`, and never a `500`.

The empty ID is the only `400`, and it never reaches an adapter: ogen treats an
empty path parameter as a missing required one. The adapters still handle `""` as
absent, because the MCP tools take `id` from their arguments with no decoder in
front of them.

`ErrDBConvertError` is reserved for stored data — a payload field that is missing
or not a string — and never describes what the caller passed in.

## Why

The alphabet belongs to one backend, so the contract cannot state it: a format rule
in `openapi.yaml` would either forbid keys the embedded store accepts or promise a
shape Qdrant cannot always honour. Answering "absent" keeps the two backends
indistinguishable for every input a caller can actually send.

An empty ID is different in kind: it is not a value the client chose but a missing
segment, and it is the same for both backends because neither is called. Reporting
it as `400` says what happened — the path is incomplete — instead of pretending a
fact was looked up.

A `400` from the adapter was rejected for the opposite reason: `abc` is a string
the contract permits, so the request is not malformed. Before this rule the same
call answered `404` on `kv` and `500 DB_CONVERT_ERROR` on Qdrant, telling a client
its storage was broken when it had only asked for something that does not exist.
