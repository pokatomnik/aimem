# One Fact Per Write Request

## Decision

The API has no bulk endpoint. A client that needs to insert many facts issues one
`POST /api/v1/fact` per fact. The former `import` command no longer exists.

## Why

A fact's identity and its duplicate rule are per text (see `fact-id-from-text.md`
and `stored-records-immutable.md`), so a per-fact request carries exactly the unit
those rules talk about, and every write goes through the same read-then-write path
with the same accepted race.

A batch request would have to answer a question a single write never faces: which
items were written when the third one is a duplicate. That needs a per-item result
shape in the contract and in the UI, for a client that can simply loop — and it
would let one request's failure land in the middle of a stored record set with no
way to describe the outcome.
