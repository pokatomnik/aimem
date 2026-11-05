# Fact ID Derivation

## Decision

A fact's ID is derived from its text, so the ID and the content cannot disagree.

The digest is shaped per backend:

- `kv` stores the full 64-character SHA-256 hex;
- Qdrant takes the first 32 hex characters as a UUID — Qdrant accepts only a UUID
  or `u64`, and its parser does not check the version or variant bits.

Both accept a full digest on lookup, so a `kv` fact resolves to the same Qdrant
point. Which strings each backend can accept at all, and how a string it cannot
accept is answered, is `id-validity-is-backend-local.md`.

## Why

Content addressing makes the duplicate check exact and free: writing the same text
again produces the same ID, which is what triggers `ErrAlreadyHasFact` (see
`stored-records-immutable.md`). A generated ID would need a lookup by content
anyway, with its own index and a wider window where two writers miss each other,
and the same text stored twice would become a silent duplicate instead of an
error.

Truncating to 32 hex characters for Qdrant costs nothing here: the ID space stays
far larger than any operator's fact count, and accepting the full digest on lookup
means one rule covers both backends.
