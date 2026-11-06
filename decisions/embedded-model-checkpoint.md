# Embedded Model Checkpoint

## Decision

The EmbeddingGemma 2 checkpoint is compiled into the binary. `internal/adapters/embeder/gemma.go`
embeds the `gemma` directory with `//go:embed gemma`, and four `//go:generate` lines fetch the
four files — `config.json`, `tokenizer.json`, `config_sentence_transformers.json` and
`model.safetensors` — before the compiler runs.

The checkpoint is not committed. `internal/adapters/embeder/.gitignore` ignores `gemma/`, so a
fresh clone downloads about 1.4 GB with `go generate ./internal/adapters/embeder/` and `go build`
fails with `pattern gemma: no matching files found` until it does. The order is always generate,
then build.

The download is pinned to revision `f9972ac22e696a105c482202646979e33e4b5a1e`, not to `main`. The
revision appears in all four URLs because `go:generate` cannot reference a Go constant; the four
lines change together or not at all.

Each line is `sh -c "[ -s <file> ] || curl -fL --create-dirs -o <file> <url>"`, so a present,
non-empty file is never re-fetched and re-running the generator costs nothing. `-f` keeps a 404
from writing an HTML error page into `config.json`; `-s` makes a zero-byte file count as missing;
`-L` follows the `/resolve/` redirect; `--create-dirs` creates `gemma/` in a fresh clone.

At runtime `ensureModel` unpacks the embedded files into `~/.aimem/models/gemma` (selected in
`scaffold`) and the loader reads them from there.

## Why

The alternative that was rejected is committing the weights: 1.4 GB of binary blobs makes every
clone and every history operation pay for data that never diffs.

The other rejected alternative is the previous behaviour — downloading at runtime on first use.
The weights are fixed, so a build-time fetch makes the delivered binary self-contained and moves
a network failure from the first request to the build, where it is immediately visible.

Pinning to a commit rather than to `main` is correctness, not hygiene. A different checkpoint
produces different vectors, and the facts already in storage were embedded with this one, so an
upstream update picked up silently would leave stored facts unreachable by similarity search.
That is a failure nobody notices until recall comes back empty.

Two costs are accepted rather than fixed: no atomic download (an interrupted `curl` can leave a
truncated file, which `[ -s ]` accepts as present — the user deletes it), and no context
cancellation (unpacking happens once, at startup, for work that cannot be abandoned halfway).
