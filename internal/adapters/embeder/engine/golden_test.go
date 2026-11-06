package engine

import (
	"embed"
	"encoding/json"
	"io/fs"
	"math"
	"testing"
)

// The fixture is goinfer's testdata/embeddinggemma2-tiny: a random-weight model
// built with the real modeling code under the real tensor names and config
// layout, together with golden.json, the reference output of transformers 5.19.0
// / torch 2.14.0 for four inputs. It is committed because a go:embed pattern
// cannot reach into the module cache, and because comparing against a reference
// implementation is the only way to catch a change in the arithmetic rather than
// just in the plumbing.
const goldenDir = "testdata/golden"

//go:embed all:testdata/golden
var goldenFS embed.FS

// golden is the subset of golden.json these tests read.
type golden struct {
	Transformers string `json:"transformers"`
	Torch        string `json:"torch"`
	Cases        []struct {
		IDs        []int     `json:"ids"`
		Pooled     []float64 `json:"pooled"`
		Normalized []float64 `json:"normalized"`
		F64        []float64 `json:"normalized_f64"`
	} `json:"cases"`
	SlidingMask [][]bool `json:"sliding_mask"`
}

func loadGolden(t *testing.T) golden {
	t.Helper()
	raw, err := goldenFS.ReadFile(goldenDir + "/golden.json")
	if err != nil {
		t.Fatalf("read golden.json: %v", err)
	}
	var g golden
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("parse golden.json: %v", err)
	}
	if len(g.Cases) == 0 {
		t.Fatal("golden.json has no cases")
	}
	return g
}

// cosMax returns the cosine similarity between got and want and the largest
// absolute difference.
func cosMax(got []float32, want []float64) (cos, maxAbs float64) {
	var dot, ng, nw float64
	for i, w := range want {
		g := float64(got[i])
		dot += g * w
		ng += g * g
		nw += w * w
		maxAbs = math.Max(maxAbs, math.Abs(g-w))
	}
	return dot / math.Sqrt(ng*nw), maxAbs
}

// TestGoldenFromFS is the gate: the checkpoint is read out of an embed.FS, with
// no file on disk, and every case must land where upstream's own test lands
// against the transformers reference (cosine >= 0.9999).
//
// It goes through load rather than LoadEncoder because the fixture ships no
// tokenizer.json: the numbers depend on the weights and the ids, not on how a
// string becomes ids, and the tokenizer half has its own test.
func TestGoldenFromFS(t *testing.T) {
	g := loadGolden(t)

	// A sub-FS reproduces production: the caller hands the loader the directory
	// that holds the checkpoint, not the FS root.
	sub, err := fs.Sub(goldenFS, goldenDir)
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}

	m, err := load(sub, ".")
	if err != nil {
		t.Fatalf("load from embed.FS: %v", err)
	}

	for i, c := range g.Cases {
		got, err := m.embedIDs(c.IDs)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		if len(got) != len(c.Normalized) {
			t.Fatalf("case %d: got %d values, want %d", i, len(got), len(c.Normalized))
		}
		cos, maxAbs := cosMax(got, c.Normalized)
		t.Logf("case %d ids=%v cos=%.9f maxAbs=%.3g", i, c.IDs, cos, maxAbs)
		if cos < 0.9999 {
			t.Errorf("case %d: cosine %.9f below 0.9999", i, cos)
		}
	}
}
