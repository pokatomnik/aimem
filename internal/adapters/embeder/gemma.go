package embeder

import "embed"

// The checkpoint is not kept in git; run `go generate ./internal/adapters/embeder/`
// to fetch it. It is pinned to a commit rather than `main` so that a rebuild
// embeds the same weights: a different checkpoint moves every vector, and the
// facts already in storage were embedded with this one.
//
//go:embed gemma
var gemmaDir embed.FS

const gemmaDirName = "gemma"

// Each line downloads its file only when the file is missing or empty, so
// re-running the generator costs nothing. The guard needs `sh -c`: go generate
// execs the line directly and does not interpret `||` itself. The revision
// appears in all four URLs because go:generate cannot reference a Go constant:
// change every occurrence together.
//
//go:generate sh -c "[ -s gemma/config.json ] || curl -fL --create-dirs -o gemma/config.json https://huggingface.co/google/embeddinggemma-2/resolve/f9972ac22e696a105c482202646979e33e4b5a1e/config.json"
//go:generate sh -c "[ -s gemma/tokenizer.json ] || curl -fL --create-dirs -o gemma/tokenizer.json https://huggingface.co/google/embeddinggemma-2/resolve/f9972ac22e696a105c482202646979e33e4b5a1e/tokenizer.json"
//go:generate sh -c "[ -s gemma/config_sentence_transformers.json ] || curl -fL --create-dirs -o gemma/config_sentence_transformers.json https://huggingface.co/google/embeddinggemma-2/resolve/f9972ac22e696a105c482202646979e33e4b5a1e/config_sentence_transformers.json"
//go:generate sh -c "[ -s gemma/model.safetensors ] || curl -fL --create-dirs -o gemma/model.safetensors https://huggingface.co/google/embeddinggemma-2/resolve/f9972ac22e696a105c482202646979e33e4b5a1e/model.safetensors"
