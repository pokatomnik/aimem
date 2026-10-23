package kv

import (
	"context"
	"testing"
)

type testConfig struct {
	dimensions uint32
}

func (c testConfig) GetDimensions() uint32 { return c.dimensions }

// A fact identical to the query must survive the similarity filter, and a
// distant one must not, regardless of how the underlying metric reports score.
func TestGetFactsListSimilarityThreshold(t *testing.T) {
	kv := New(testConfig{dimensions: 2}, t.TempDir())

	ctx := context.Background()

	facts := []struct {
		vector  []float32
		content string
	}{
		{[]float32{1, 0}, "near"},
		{[]float32{0, 1}, "far"},
	}

	for _, f := range facts {
		if err := kv.AddFact(ctx, f.vector, "summary "+f.content, f.content); err != nil {
			t.Fatal(err)
		}
	}

	// A high threshold must keep only the near-identical fact.
	near, err := kv.GetFactsList(ctx, []float32{1, 0}, 10, 0.9)
	if err != nil {
		t.Fatal(err)
	}

	if len(near) != 1 {
		t.Fatalf("expected 1 fact above threshold 0.9, got %d: %+v", len(near), near)
	}
	if near[0].Summary != "summary near" {
		t.Errorf("wrong fact survived: %q", near[0].Summary)
	}

	// A zero threshold must keep everything.
	all, err := kv.GetFactsList(ctx, []float32{1, 0}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected both facts at threshold 0, got %d", len(all))
	}

	// A threshold of 1 keeps only an exact match.
	exact, err := kv.GetFactsList(ctx, []float32{1, 0}, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(exact) != 1 {
		t.Fatalf("expected only the exact match at threshold 1, got %d: %+v", len(exact), exact)
	}
	if exact[0].Summary != "summary near" {
		t.Errorf("wrong fact survived at threshold 1: %q", exact[0].Summary)
	}
}

// TopK is an upper bound applied before filtering, so a filtered result may be
// shorter than TopK, including empty.
func TestGetFactsListRespectsTopK(t *testing.T) {
	kv := New(testConfig{dimensions: 2}, t.TempDir())

	ctx := context.Background()

	for i, v := range [][]float32{{1, 0}, {1, 0.1}, {1, 0.2}} {
		if err := kv.AddFact(ctx, v, "summary", string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}

	got, err := kv.GetFactsList(ctx, []float32{1, 0}, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) > 2 {
		t.Errorf("TopK=2 returned %d facts", len(got))
	}
}
