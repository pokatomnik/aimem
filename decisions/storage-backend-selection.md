# Storage Backend Selection

## Decision

Two backends implement one storage contract: the embedded `xvec` index in
`internal/repositories/kv`, and Qdrant over HTTP in `internal/repositories/qdrant`.
Both satisfy the interfaces the use cases declare.

`scaffold` picks Qdrant when `AIMEM_QDRANT_HOST` is non-empty, and the embedded
backend otherwise.

Runtime settings come from `AIMEM_*` process environment variables, read by
`cmd/aimem/config.go`. There are no CLI flags, no `.env` loader and no config
file: the name, whether a setting is required, and its default all live in the
`Config` env tags. `AIMEM_EMBEDDING_MODEL`, `AIMEM_SUMMARIZATION_MODEL` and
`AIMEM_DIMENSIONS` are required. The Qdrant user list travels in
`AIMEM_USERS_JSON` like every other setting.

## Why

One contract with two implementations is what keeps a local single-operator setup
and a shared server from drifting: a use case cannot tell which backend it got, so
a change made for Qdrant has to keep the embedded path honest.

Selecting on an empty host rather than on a separate `AIMEM_STORAGE` switch leaves
no third state to handle — unset is the local default, set is a server — and one
variable cannot contradict another.

The environment is the config surface because it is what the deployment already
provides, and the struct tags keep each setting's name, whether it is required, and
its default in one place, instead of spreading them across a template, a loader and
the docs.
