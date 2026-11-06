---
name: gemma-engine-sync
description: Takes upstream changes by hand into the vendored EmbeddingGemma 2 text engine in internal/adapters/embeder/engine, which is a modified copy of goinfer's embeddinggemma2/model.go and encoder.go. Classifies each upstream change as arithmetic (every stored vector becomes unreachable — a breaking change that forces the next major release) or as loading only, and then applies it without losing the seven local adaptations. Use when goinfer publishes a new version, when the embedding engine must be updated or a tensor name fails to load, or when someone proposes editing engine/model.go for any reason.
---

# Gemma Engine Sync

`internal/adapters/embeder/engine` is a copy of upstream code, not a port that
tracks it. Nothing brings changes in automatically, and nothing can: the package
exists precisely because upstream cannot read weights from anything but a real
file. This skill is the whole maintenance procedure, and its most important job is
to decide, before a single line changes, whether the update moves the numbers.

`decisions/gemma-engine-vendored.md` records why the copy exists. Read it first if
the question is "should we still be doing this"; read this file if the question is
"how do we take the change".

## The one rule that outranks the rest

**A change that alters the arithmetic invalidates every vector already in
storage.** A stored fact is found again by cosine similarity against the vector it
was written with. Embed the same text with different numbers and the two vectors
land in unrelated places: recall returns nothing, and nothing errors. This is the
failure mode the checkpoint pinning in `embedded-model-checkpoint.md` exists to
prevent, and it applies to code exactly as it applies to weights.

So: an update that changes the numbers is a **breaking change of this project**.
The release that carries it must be the next **major** version — not a minor, not a
patch — and its notes must say that existing facts have to be re-embedded. Do not
ship an arithmetic change under any other version, and do not merge one without the
re-embedding path being decided first.

Everything below exists to answer that question before it is too late to ask.

## What is vendored, and what is not

| Upstream | Here | Note |
|---|---|---|
| `embeddinggemma2/model.go` | `engine/model.go` | adapted; arithmetic verbatim |
| `embeddinggemma2/encoder.go` | `engine/encoder.go` | reduced to `LoadEncoder` + `EmbedText` |
| `aikit/embed/safetensors.go` | `engine/safetensors.go` | header rules copied; read path rewritten |
| `embeddinggemma2/vision.go`, `audio.go`, `*_accel.go`, `preprocess.go` | not vendored | other modalities and accelerators |
| `tokenizer` | dependency | `LoadJSONBytes` only |
| `aikit/linalg` | dependency | `MatmulBT` only |

Do not copy more of upstream in "while we are here". Each additional file is a
file that has to be re-diffed forever, and the package's value is that it is small.

## The seven local adaptations

Any sync must end with all seven still true. They are listed in the header of
`engine/model.go` too; if a later edit changes them, change both places.

1. Imports use `io/fs` and `path`, never `os` or `path/filepath`.
2. `loadConfig(fsys, name)` reads `config.json` through `fs.ReadFile`.
3. `load(fsys, dir)` replaces upstream's `Load(dir)`.
4. `openSafetensors` (this package) replaces `embed.OpenSafetensorsMmap`.
5. No per-tensor `append([]float32(nil), v...)` copy: this reader owns its memory
   already, and the copy would double the load-time peak.
6. Exactly one exported function (`LoadEncoder`) and one exported type (`Encoder`).
7. `config_sentence_transformers.json` is not read: `EmbedText` is the no-prompt
   path, which upstream reaches by passing `""`.

## Procedure

### 1. Record where we are

```sh
git --no-pager log -1 --format='%h %s' -- internal/adapters/embeder/engine
grep -n "goinfer v" decisions/gemma-engine-vendored.md internal/adapters/embeder/engine/model.go
```

The version in the decision record is the baseline. If it disagrees with the
`goinfer` requirement in `go.mod`, that disagreement is the first thing to
understand — do not proceed until it is resolved.

### 2. Read the upstream change

Upstream is `github.com/townsendmerino/goinfer`, branch **`main`**. Files:
`embeddinggemma2/model.go`, `embeddinggemma2/encoder.go`, and — only if the header
rules are in question — `aikit/embed/safetensors.go` from
`github.com/townsendmerino/aikit`.

The diff against our copy is noisy by construction: names and the read path differ
everywhere. The diff that decides the classification is **upstream against
upstream** — the release we vendored versus `main` — so that our adaptations are
not mistaken for upstream's changes. Fetch both and compare:

```sh
curl -fsSL https://raw.githubusercontent.com/townsendmerino/goinfer/main/embeddinggemma2/model.go -o /tmp/up-model.go
curl -fsSL https://raw.githubusercontent.com/townsendmerino/goinfer/main/embeddinggemma2/encoder.go -o /tmp/up-encoder.go

# The vendored baseline is also on disk in the module cache.
go mod download github.com/townsendmerino/goinfer@v0.24.1
diff -u "$(go env GOMODCACHE)/github.com/townsendmerino/goinfer@v0.24.1/embeddinggemma2/model.go" /tmp/up-model.go
```

If both diffs are empty, there is nothing to take; stop.

### 3. Classify every hunk

Walk the upstream-to-upstream diff hunk by hunk and put each one in one of three
buckets. Do not skip this: a hunk that looks like cleanup can move a number.

| Bucket | What it looks like | Effect |
|---|---|---|
| **Arithmetic** | `forward`, `forwardEmbeds`, `attention`, `softmaxWindow`, `mlp`, `geluTanh`, `rmsRows`, `rmsNorm`, `ropeTables`, `rope`, `poolNormalize`, `embedTokens`, tensor names or shapes, `embedding_projection` use, pooling or normalisation order | **Breaking.** Changes every stored vector. Next major release, re-embedding, new golden. |
| **Loading** | how bytes are read, header parsing, dtype widening, validation | Not breaking **only if** it does not change a byte that reaches a weight. Verify with the golden test: it must still pass unchanged. |
| **API / comments** | renames, doc strings, exported surface, config validation messages | Not breaking. Adapt to the local naming. |

`Config` parsing is arithmetic by default: a changed `SlidingWindow`, `RMSNormEps`
or `LayerTypes` reading changes attention or normalisation. Treat "not sure" as
arithmetic.

### 4. Apply, preserving the seven adaptations

Copy upstream's arithmetic hunks verbatim. For a loading hunk, rewrite it onto
`fs.FS`/`io.ReaderAt` rather than pasting `os.Open`. For an API hunk, keep the
local rule: only `LoadEncoder` and `Encoder` are exported.

If a new tensor appears in `load` but not in the checkpoint we pin, the load fails
with that tensor's name. That is a checkpoint change and it goes through
`embedded-model-checkpoint.md` (pin a new revision, re-embed), not through this
skill.

### 5. Update the record

- The provenance header in `engine/model.go` and `engine/encoder.go`: the version
  string, and the numbered list if a difference was added or removed.
- `decisions/gemma-engine-vendored.md`: the version, and the memory numbers if the
  read path changed.

Neither file may claim something the code does not do; that is the failure
`prompt-docs-keeper` exists to catch.

### 6. Verify

```sh
go test ./internal/adapters/embeder/engine/ -v
go test ./internal/adapters/embeder/ -run TestProductionCheckpointFromBinary
make test
```

What each proves:

- **`TestGoldenFromFS`** — the only test that compares against another
  implementation (transformers 5.19.0 through upstream's committed fixture). Cosine
  must stay `>= 0.9999` per case. If it moved, the change is arithmetic: go back to
  step 3, and do not "fix" the test.
- **`TestLoadEncoderFromFS`, `TestEmbedTextIsTokenizeThenEmbed`** — the wiring:
  model plus tokenizer plus the BOS/EOS wrap.
- **`TestCheckpointFileOutlivesEveryTensorRead`, `TestClosedFileIsNotReopened`,
  `TestTensorSurvivesClose`** — the lifecycle. These exist because a draft closed
  the checkpoint inside `openSafetensors` and still passed the golden test:
  `embed.FS` and `fstest.MapFS` do not punish a premature `Close`; only a real
  descriptor does. Keep them.
- **`safetensors_test.go`** — the header rules: shape against byte range, negative
  dims, offsets past the payload, a truncated header, an unsupported dtype, a file
  without `io.ReaderAt`.
- **`TestProductionCheckpointFromBinary`** — the shipped 1.49 GB checkpoint loads
  from the binary. This is the test that notices a `//go:embed` change.

### 7. Report the classification

In the commit message, state the bucket. For a loading-only change: "vendored
engine: sync goinfer <old>..<new>, loading only, vectors unchanged (golden
unchanged)". For an arithmetic change, say so plainly and name the release as
major:

```
- Sync vendored engine to goinfer <new>
- BREAKING: embeddings change; stored facts must be re-embedded

Refs: decisions/gemma-engine-vendored.md
```

## Failure modes this skill is here to prevent

- **Editing `engine/model.go` for a local reason.** That file is upstream's
  arithmetic. A local fix belongs beside it, not inside it, or the next diff becomes
  unreadable and the arithmetic silently becomes ours to defend.
- **Passing the golden test after loosening it.** Cosine dropping to 0.9998 means
  the numbers moved. Lower the bar only if upstream lowered it, and say so.
- **Copying vision, audio or accelerator code.** Not used, not wanted.
- **Letting `engine` import `embeder` or any other layer.** The package must stay a
  leaf: standard library, `goinfer/tokenizer`, `aikit/linalg`. `scaffold.go` is what
  wires it.
- **Assuming an empty upstream diff means an empty obligation.** A new upstream
  version can move `aikit` under us. Check the `aikit` requirement too when the
  `goinfer` one moves.
