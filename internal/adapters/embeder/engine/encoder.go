package engine

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/townsendmerino/goinfer/tokenizer"
)

// This file replaces embeddinggemma2/encoder.go from goinfer v0.24.1 (MIT,
// Copyright (c) 2026 Francis Townsend-Merino; see LICENSE beside this file).
// Upstream's LoadEncoder takes a directory path and composes Load(dir),
// tokenizer.Load(dir) and os.ReadFile(dir/config_sentence_transformers.json).
// Here the checkpoint is read through fs.FS instead, and two of those three steps
// are replaced by their bytes-based twins:
//
//   - tokenizer.Load(dir) becomes tokenizer.LoadJSONBytes(fs.ReadFile(tokenizer.json)),
//     which parses the same self-contained tokenizer.json and is equivalent for
//     this checkpoint: the directory variant only opens a sibling
//     tokenizer_config.json for special tokens, and the four embedded files do
//     not include one, so the sibling lookup resolves to nothing either way.
//   - config_sentence_transformers.json is not read at all. It carries only the
//     named task-prompt table, which EmbedText reaches through the empty prompt
//     name, and PromptText returns "" for that name before consulting the table.
//
// Tokenize is upstream's, minus the prompt argument: BOS, the text, EOS, and the
// same MaxTokens refusal.

// maxTokens is the context the model card gives for every modality (8,192). An
// input longer than this is refused, not cut: a silently truncated input would
// embed a different text than the caller sent.
const maxTokens = 8192

type (
	// Encoder is an EmbeddingGemma 2 text encoder with its tokenizer. It is safe
	// for concurrent use: the weights are immutable after LoadEncoder.
	Encoder struct {
		m   *model
		tok *tokenizer.Tokenizer
		bos int
		eos int
	}
)

// LoadEncoder loads the text encoder from an HF checkpoint layout reachable
// through f: dir/config.json, dir/model.safetensors and dir/tokenizer.json. The
// vision and audio tensors in model.safetensors are not read, so a checkpoint
// carrying them loads unchanged.
//
// Nothing is written anywhere: the files are opened through f and released
// before LoadEncoder returns.
func LoadEncoder(f fs.FS, dir string) (*Encoder, error) {
	m, err := load(f, dir)
	if err != nil {
		return nil, err
	}

	raw, err := fs.ReadFile(f, path.Join(dir, "tokenizer.json"))
	if err != nil {
		return nil, fmt.Errorf("embeddinggemma2: tokenizer: %w", err)
	}
	tok, err := tokenizer.LoadJSONBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("embeddinggemma2: tokenizer: %w", err)
	}

	bos, okB := tok.TokenID("<bos>")
	eos, okE := tok.TokenID("<eos>")
	if !okB || !okE {
		return nil, fmt.Errorf("embeddinggemma2: tokenizer has no <bos>/<eos>")
	}

	return &Encoder{m: m, tok: tok, bos: bos, eos: eos}, nil
}

// EmbedText is the sentence embedding of text with no task prompt, unit length.
// It is the encoding this project stores: the same text always produces the same
// vector for a given checkpoint.
func (e *Encoder) EmbedText(text string) ([]float32, error) {
	ids, err := e.tokenize(text)
	if err != nil {
		return nil, err
	}
	return e.m.embedIDs(ids)
}

// tokenize is upstream's Encoder.Tokenize with the prompt name fixed to "": the
// ids the model sees for text are <bos>, the text, <eos>, as the checkpoint's
// tokenizer post-processor writes them. sentence-transformers mean-pools over all
// of them, both special tokens included.
func (e *Encoder) tokenize(text string) ([]int, error) {
	ids, err := e.tok.Encode(text, false)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(ids)+2)
	out = append(out, e.bos)
	out = append(out, ids...)
	out = append(out, e.eos)
	if len(out) > maxTokens {
		return nil, fmt.Errorf("input is %d tokens, over the model's %d-token context", len(out), maxTokens)
	}
	return out, nil
}
