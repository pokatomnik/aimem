// Package qdrant stores aimem's facts in Qdrant while keeping the contract of the
// embedded kv repository: the same four operations, the same collection schema,
// and the same meaning of similarity.
//
// # Storage is partitioned per user
//
// Every operation takes a userID and acts only on that user's collection, named
// from it (see collectionName). The partition is a collection per user rather
// than a field on a shared one, so a fact belonging to someone else is not merely
// filtered out — it does not exist in the caller's collection, and the same text
// stored by two users stays two independent records.
//
// kv ignores the same userID: the embedded backend is local files in a
// single-user deployment. The two share the public contract, not the storage
// layout.
//
// # Records are immutable
//
// A fact's identity is the SHA-256 digest of its text, reduced to a UUID, so the
// ID and the content always agree. No operation changes a stored fact in place:
// writing the same text again is answered with apperrors.ErrAlreadyHasFact
// instead of overwriting it.
//
// This is a hard rule for every write to the collection, not a property of the
// current call sites. A collection here is always one user's collection. A
// stored point is a complete record, and it is never patched. If a fact ever has
// to change, the change is two steps — delete the old point, then insert the new
// one — which keeps every stored point complete and self-consistent and removes a
// whole class of partial-write problems: no half-updated payload, no vector and
// payload from different writes, no question of which field the last writer
// touched.
//
// The package therefore touches only four endpoints: insert, retrieve, query and
// delete. The payload- and vector-update endpoints are deliberately unused, and
// a test asserts that no request ever reaches them.
//
// # Known limitation
//
// AddFact checks for an existing point and then inserts, in two separate
// requests. Two concurrent writers of the same text, in the same user's
// collection, can both pass the check, and the second insert then replaces the
// record: same ID, same content, a fresh created_at. The data stays correct
// because the ID is derived from the content and both writers hold the same fact,
// but the loser sees success rather than ErrAlreadyHasFact.
//
// Closing the window would need a conditional insert. Qdrant keeps one
// internally (conditional_upsert) but does not expose it over REST or gRPC in the
// pinned version, and its update_filter guards updates of existing points rather
// than refusing an insert. A process-local lock is not a fix: a second instance
// of the application is a separate process and would not share it, so the lock
// would only hide the window behind a false sense of safety. Writes here are far
// rarer than reads, so the window is accepted and documented rather than patched.
package qdrant
