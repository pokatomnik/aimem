# Stored Records Are Immutable

## Decision

A fact is written once and never patched in place. This holds for `kv` and
`qdrant` alike, and for any backend added later. Concretely: no in-place update
APIs — no payload update, no vector update, no partial or batch overwrite.

Writing the same text again answers `ErrAlreadyHasFact` instead of overwriting. A
change, if one is ever needed, is two explicit steps: delete the record, then
write the new one. A stored record is therefore always the result of exactly one
write.

`internal/repositories/qdrant/immutability_test.go` enforces this for the Qdrant
backend by failing if any operation reaches a mutation endpoint.

The write path is read-then-write: `AddFact` checks whether the point exists and
then inserts it, in two requests. Two concurrent writers of the same text can both
pass the check; the loser rewrites the same ID with the same content and sees
success instead of `ErrAlreadyHasFact`. That window is accepted — do not paper
over it with a process-local lock.

## Why

One write per record makes a read unambiguous: there is no half-updated payload to
observe, no version to compare, and no way for stored content to disagree with the
ID derived from it.

`ErrAlreadyHasFact` is the point, not an obstacle: overwriting silently would let a
re-run of any client erase a fact the user still relies on.

The race is accepted because the data stays correct — the loser writes the same ID
with the same content — and because writes are far rarer than reads. A
process-local lock would not close it anyway: a second instance is a separate
process and would not share the lock, so the lock would only hide the window while
making the code look safe.
