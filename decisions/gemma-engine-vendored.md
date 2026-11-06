# EmbeddingGemma 2 Text Forward, Vendored

## Decision

The text half of EmbeddingGemma 2 is computed by code that lives in this
repository: `internal/adapters/embeder/engine`. It is a copy of
`embeddinggemma2/model.go` and `embeddinggemma2/encoder.go` from
`github.com/townsendmerino/goinfer` v0.24.1, adapted so the checkpoint is read
through `fs.FS` — from `//go:embed`, out of the binary — instead of from a
directory on disk.

The package exports exactly one constructor, `LoadEncoder(f fs.FS, dir string)`,
and one type, `Encoder`, whose only method is `EmbedText(text string)`. The model
arithmetic is upstream's and is not to be edited here; the vendored copy differs
only in how bytes reach it. The seven differences are listed in the file header and
repeated in the pull request that adds this package.

The weights stay embedded whole (see `model-checkpoint-not-trimmed.md`) and are
never unpacked. Nothing is written to `~/.aimem`, or anywhere else, on the path
from binary to embedding.

`goinfer/tokenizer` and `aikit/linalg` remain dependencies: the first parses
`tokenizer.json` (`LoadJSONBytes`, upstream's stable bytes-based entry point) and
the second provides `MatmulBT`. Neither is copied.

Changes are taken from upstream by hand, following
`.agents/skills/gemma-engine-sync`. That skill is the only sanctioned procedure,
and it treats a change that moves the numbers as a breaking change of this project.

## Why

The library can only open a checkpoint from a real file: `embeddinggemma2.Load`
calls `aikit`'s `OpenSafetensorsMmap`, which is `os.Open` plus `syscall.Mmap`, and
`*Model` — the value that actually reads the weights — cannot be constructed from
outside its package: its fields are unexported, `Load` is the only constructor,
`Weights()` is one-way, and `RegisterAccelerator` hands it out only from a path that
already went through disk. There is no `LoadFromFS` in any published version.
Reading the weights out of the binary therefore requires owning the code that reads
them.

The alternatives were rejected on their merits. Unpacking into a temporary
directory removes the permanent copy but writes 1.49 GB at every start and keeps a
second failure mode. A `replace`-based fork carries the whole module, test fixtures
included. Reconstructing `*Model` through `unsafe`/`reflect` breaks silently on any
upstream field reordering. OS-level tricks (FUSE, `LD_PRELOAD`, RAM disks,
`memfd` plus `/proc/self/fd`) need privileges, do not exist on macOS, or cost the
same memory without the portability.

Vendoring is affordable here for a reason specific to this project: the embedding
engine sits behind a two-method interface declared by the use cases, so it is
replaceable. If a better engine — a hosted API, another local model — arrives, the
package is deleted and the interface is satisfied elsewhere. That is also why the
maintenance cost of tracking upstream by hand is acceptable rather than permanent.

The copy also removes two defects the wrapper had. `ensureModel` ran on every
`GetEmbeddings` call rather than once, and it validated a cached file by size
alone, so a rebuilt binary kept reading the previous checkpoint. Neither exists
when the weights come from the binary that is running.

Measured on the checkpoint this project pins: peak RSS while loading drops from
1.83 GB to 1.67 GB, because upstream copies each tensor a second time
(`append([]float32(nil), v...)`) to detach it from a mapping it is about to close,
and this reader always decodes into memory it owns.

Licensing: both upstream modules are MIT. The copyright notice is kept in
`internal/adapters/embeder/engine/LICENSE` and referenced from the file headers.
