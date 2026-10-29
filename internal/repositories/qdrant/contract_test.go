package qdrant

import (
	"context"
	"testing"

	"github.com/pokatomnik/gomem/internal/entities/fact"
)

// store is the contract the use cases consume, copied verbatim from the
// interfaces they declare in internal/use_cases. A *Qdrant must satisfy it, which
// is what makes this repository a drop-in replacement for kv.
type store interface {
	AddFact(ctx context.Context, userID string, vector []float32, summary, content string) error
	GetFactByID(ctx context.Context, userID, factID string) (fact.FactFullView, error)
	GetFactsList(ctx context.Context, userID string, vector []float32, topK int, similarityThreshold float32) (facts []fact.FactListView, err error)
	RemoveFactByID(ctx context.Context, userID, factID string) error
}

// The assertion is the test: it fails to compile if a signature drifts from the
// one kv exposes.
var _ store = (*Qdrant)(nil)

// New must hand back a pointer, so callers share one repository.
var _ *Qdrant = New(fakeConfig{}, nil)

func TestNewFallsBackToDefaultClient(t *testing.T) {
	t.Parallel()

	repository := New(fakeConfig{host: "http://qdrant.test"}, nil)
	if repository.client == nil {
		t.Fatalf("New(nil client) left client nil")
	}
}
