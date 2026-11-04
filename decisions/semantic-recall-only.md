# Semantic Recall Only

## Decision

Recall is semantic only and returns summaries. There are no tags, no filters and
no keyword search.

`topK` is applied first; the results are then filtered by `similarityThreshold`. A
result set can therefore come back empty even when similar facts exist.

The two backends report the score differently, and each adapter must respect it:

- `xvec` returns a cosine **distance** (0 = identical), so `kv` converts it with
  `1 - result.Score`;
- Qdrant's cosine score is already a **similarity** and must not be inverted.

## Why

Every selector recall accepted would have to be stored, indexed and answered
identically by two backends, for a store that is one operator's memory. The
embedding already carries the meaning, so an extra index buys little and doubles
the surface that can disagree between `kv` and Qdrant.

Summaries are what recall returns because a summary is what `remember` stores, so
a caller that needs the exact wording asks for the full fact by ID — a deliberate
second step, and the reason the two tools are separate.

The score direction is the one place the backends genuinely differ, and getting it
backwards inverts the filter: written as `if score < threshold { continue }`
against an unconverted xvec distance, every close result is dropped and only the
distant ones survive.
