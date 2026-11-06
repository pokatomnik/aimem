package embeder

import (
	"math"
	"testing"

	"github.com/pokatomnik/gomem/internal/adapters/embeder/engine"
)

// TestProductionCheckpointFromBinary is the end-to-end check of the reason the
// engine package exists: the real 1.49 GB checkpoint is read out of the binary's
// read-only data through embed.FS, with no file on disk and no unpacking step.
//
// It is hermetic by the project's definition — embed.FS is resolved by the
// compiler, so nothing is read from disk, no socket is opened and no environment
// is consulted — but it is heavy: it materialises the float32 weights (about
// 1 GB) and takes a second or so. `go test -short` skips it; the numbers
// themselves are covered hermetically by engine's golden test, so what this adds
// is proof that the shipped checkpoint loads at all.
func TestProductionCheckpointFromBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("loads the full checkpoint into memory")
	}

	enc, err := engine.LoadEncoder(gemmaDir, gemmaDirName)
	if err != nil {
		t.Fatalf("load the embedded checkpoint: %v", err)
	}

	const text = "The quick brown fox jumps over the lazy dog"
	first, err := enc.EmbedText(text)
	if err != nil {
		t.Fatalf("EmbedText: %v", err)
	}
	if len(first) != 768 {
		t.Fatalf("embedding width %d, want 768", len(first))
	}

	var sumSquares float64
	for _, v := range first {
		sumSquares += float64(v) * float64(v)
	}
	if norm := math.Sqrt(sumSquares); math.Abs(norm-1) > 1e-3 {
		t.Fatalf("embedding is not unit length: norm %v", norm)
	}

	// The same text must map to the same vector on every call: a stored fact is
	// found again by similarity, so a non-deterministic encoder silently breaks
	// recall.
	second, err := enc.EmbedText(text)
	if err != nil {
		t.Fatalf("EmbedText again: %v", err)
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("value %d differs between calls: %v != %v", i, first[i], second[i])
		}
	}
}
