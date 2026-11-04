# Per-User Collections in Qdrant

## Decision

Users are separated by collection in Qdrant, not by a payload field. All four
operations take a `userID` and act on that user's collection, named by
`collectionName`: `memory-<userID>`, or the bare `memory` for an empty one.

A `userID` is an opaque identifier. Because a collection name becomes part of the
URL path, Qdrant allows only ASCII letters, digits, `_` and `-`; a `/` would
address a different route.

`kv` takes the same `userID` and ignores it on purpose. The backends share the
public contract, not the storage layout.

## Why

The point ID comes from the text alone (see `fact-id-from-text.md`), so in one
shared collection two users storing the same text would collide, and the second
`AddFact` would answer `ErrAlreadyHasFact` for a fact the caller never wrote.

A separate collection also makes a foreign fact simply absent: `GetFactByID`
returns `ErrFactDoesNotExist` rather than a permission error, and no ownership
check is needed on any path.

Partitioning by a payload field was rejected: it would need a filter on every read
and a matching index, and the retrieve endpoint accepts no filter at all, so a
lookup by ID could not be scoped to the caller.
