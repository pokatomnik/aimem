# The Model Checkpoint Is Read From the Binary

## Decision

The embedding checkpoint is never written to disk. `internal/adapters/embeder`
embeds the `gemma` directory with `//go:embed`, and `engine.LoadEncoder` reads
`config.json`, `tokenizer.json` and `model.safetensors` straight out of the
resulting `embed.FS` through the `fs.FS` interface the vendored engine takes.

Nothing is unpacked into `~/.aimem/models/gemma`, or anywhere else, on the path
from binary to embedding. There is no model directory to keep, validate, delete
or restore, and both storage backends behave identically in this respect.

This supersedes the runtime-unpacking step in
[`embedded-model-checkpoint.md`](embedded-model-checkpoint.md). Everything else
in that record still holds: the checkpoint is compiled into the binary, fetched
at build time, and pinned to one revision.

## Why

Upstream's loader can only open a checkpoint from a real file, so the choice was
between materialising 1.49 GB on every start and owning the code that reads the
weights. Reading through `fs.FS` removes the copy, the startup cost and a second
failure mode at once.

An unpacked copy was also a second, quieter source of truth. A stale directory and
a rebuilt binary could disagree about which weights are in use, and the
disagreement would surface as vectors that load and rank slightly wrong rather
than as an error. The unpacking step that was removed had exactly that defect: it
validated a cached file by size alone, so a rebuilt binary kept reading the
previous checkpoint, and it ran on every embedding call instead of once.

Keeping the files on disk was rejected for the same reason the checkpoint is
pinned: the weights are fixed, and one copy of a fixed input is enough.
