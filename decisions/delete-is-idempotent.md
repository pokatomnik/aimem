# Deleting a Missing Fact Is a No-Op

## Decision

`RemoveFactByID` succeeds for any ID that is not stored. `DELETE
/api/v1/fact/{factID}` answers `200`, and the MCP tool `remove_fact_by_id` reports
`ok: true`. An absent fact is never an error.

This holds for both backends and is not a property of either store. `kv` reads
xvec's `ErrNotFound` — both the batch error and the per-key one — as success;
`qdrant` uses the delete endpoint, which accepts a point that is not there. Neither
checks existence first; the check that used to answer `404` is gone.

An empty ID never reaches the repository over HTTP: ogen's parameter decoder treats
an empty path segment as a missing required parameter and answers `400` before any
handler runs. The rule above still applies to the repository, because the MCP tool
`remove_fact_by_id` takes `id` from its arguments with no such gate.

Reporting absence belongs to the read: `GET /api/v1/fact/{factID}` answers `404`
`FACT_DOES_NOT_EXIST`.

## Why

DELETE is idempotent in HTTP: a client that retries after a timeout must not see
the second attempt fail because the first one worked. An error on an absent fact
turns every retry into a false failure, and the distinction is useless — in either
outcome the fact is gone.

`openapi.yaml` already said so: `removeFactByID` declares `200` and `401` only,
while `getFactByID` declares `404`. The implementation disagreed with its own spec
because it checked existence first, to reproduce `kv`'s per-key error. The contract
was right; the code was not.

Read and delete answer differently because they promise different things: a read
that finds nothing has nothing to show, while a delete that finds nothing has
achieved what it was asked.
