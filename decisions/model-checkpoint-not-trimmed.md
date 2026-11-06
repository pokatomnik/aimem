# Model Checkpoint Not Trimmed

## Decision

The checkpoint is embedded whole, at about 1.4 GB. It is not reduced to the text-only part.

The full file carries three sets of towers. Only `language_model.*` is ever read:
`embeddinggemma2.Load` (`goinfer@v0.24.1/embeddinggemma2/model.go:128`) fetches tensors by that
prefix, and the vision and audio towers load only through `EnableVision`/`EnableAudio`, which
this project never calls. A text-only checkpoint would be about 542 MB — `audio_tower` is 609.7 MB
and `vision_tower` 334.7 MB — a 2.75x reduction.

Dropping the unread tensors is byte-exact: the kept ranges are copied verbatim and only their
`data_offsets` are recalculated, so the embeddings are identical. This was measured, not assumed:
a stripped file produced a maximum difference of zero across three texts, Cyrillic included.

## Why

The build already has one generation step that can fail before the compiler runs. A program that
rewrites the weights would be a second one, and unlike `ogen` it would produce a 1.4 GB artifact
whose correctness is invisible: a wrong offset or a short write surfaces as a model load error, or
worse, as vectors that load and rank slightly wrong. Two failure points on the path to a working
binary is one more than is worth paying for a build artifact nobody ships yet.

The saving is also unrealized today. A slimmer checkpoint pays off when the binary is distributed;
here it is built and run on one machine, where 1.4 GB is an inconvenience, while the costs — an
unverified derivative, a writer of our own, a second generation step — are paid on every build.

Trimming is a reasonable future improvement, and the approach is settled: a `//go:build ignore`
program beside the package, run from `//go:generate`, keeping the original in the user cache
rather than in `gemma/` (where `go:embed gemma` would swallow it too), verifying SHA-256 of both
input and output, and padding the header to a multiple of 8 so bf16 tensors stay aligned for
`aikit`'s `Uint16s`. Revisit it when the binary starts being distributed, not before.

Half of what would remain is `embed_tokens.weight`, so nothing below 542 MB is reachable without
quantization or vocabulary surgery — both change the numbers and need their own evaluation.
