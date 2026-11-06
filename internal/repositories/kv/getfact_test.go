package kv

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gorse-io/xvec"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

type (
	testConfig struct{}

	// fakeCollection answers Fetch and Delete the way xvec does.
	//
	// For a primary key it does not hold, Fetch returns a slice as long as the
	// key batch with a nil document for the missing key, so a length check never
	// sees "not found". Delete is worse: it returns a non-empty result list *and*
	// a batch error, with the same cause repeated in the per-key entry.
	fakeCollection struct {
		fetched []*xvec.Document

		// deleted and deleteErr are returned together, mirroring xvec.
		deleted   []xvec.WriteResult
		deleteErr error
	}
)

func (testConfig) GetDimensions() uint32 { return 2 }

func (f *fakeCollection) Insert(context.Context, []xvec.Document) ([]xvec.WriteResult, error) {
	return nil, nil
}

func (f *fakeCollection) Query(context.Context, xvec.VectorQuery) ([]xvec.Document, error) {
	return nil, nil
}

func (f *fakeCollection) Fetch(context.Context, []string, xvec.Projection) ([]*xvec.Document, error) {
	return f.fetched, nil
}

func (f *fakeCollection) Delete(context.Context, []string) ([]xvec.WriteResult, error) {
	return f.deleted, f.deleteErr
}

func (f *fakeCollection) Close() error { return nil }

// newFakeKV builds a KV whose collection is the fake, so no database is opened.
func newFakeKV(f *fakeCollection) *KV {
	kv := New(testConfig{}, "", testLogger())

	kv.open = func(context.Context) (collection, error) {
		return f, nil
	}

	return kv
}

// A missing fact must be reported as ErrFactDoesNotExist. xvec returns a slice
// with one nil entry for an absent key, so a len(result) == 0 check is dead
// code and result[0].Fields dereferences nil — on darwin/arm64 that does not
// panic, it spins forever, and the HTTP handler never answers.
func TestGetFactByIDMissingFact(t *testing.T) {
	kv := newFakeKV(&fakeCollection{fetched: []*xvec.Document{nil}})

	_, err := kv.GetFactByID(context.Background(), "", "absent")

	if !errors.Is(err, apperrors.ErrFactDoesNotExist) {
		t.Fatalf("want ErrFactDoesNotExist, got %v", err)
	}
}

// A stored fact keeps the exact shape the domain expects.
func TestGetFactByIDPresentFact(t *testing.T) {
	kv := newFakeKV(&fakeCollection{fetched: []*xvec.Document{{
		PrimaryKey: "present",
		Fields: map[string]any{
			keyFieldCreatedAt: "2026-01-01T00:00:00Z",
			keyFieldSummary:   "a summary",
			keyFieldFact:      "the full text",
		},
	}}})

	view, err := kv.GetFactByID(context.Background(), "", "present")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if view.ID != "present" {
		t.Errorf("ID = %q, want %q", view.ID, "present")
	}

	if view.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q", view.CreatedAt)
	}

	if view.Summary != "a summary" {
		t.Errorf("Summary = %q", view.Summary)
	}

	if view.Content != "the full text" {
		t.Errorf("Content = %q", view.Content)
	}
}

// Removing a fact that is not stored is a no-op, not an error: DELETE is
// idempotent and a client must be able to retry it. xvec reports the miss twice
// over — as a batch error wrapping ErrNotFound and as the same cause in the
// per-key result — and both mean "nothing to delete". The fake reproduces both
// shapes, because a check that only looks at one of them would still leak.
//
// Every error wraps xvec.ErrNotFound the way the real ones do. Building them
// with errors.New would only copy the message text, and errors.Is compares
// identity or an Is method, never the string.
func TestRemoveFactByIDMissingFact(t *testing.T) {
	cause := fmt.Errorf("xvec: DELETE: db: primary key not found: %q: %w", "absent", xvec.ErrNotFound)
	batch := fmt.Errorf("xvec: batch write: %d document operations failed: %w", 1, xvec.ErrNotFound)

	kv := newFakeKV(&fakeCollection{
		deleted:   []xvec.WriteResult{{PrimaryKey: "absent", Err: cause}},
		deleteErr: batch,
	})

	if err := kv.RemoveFactByID(context.Background(), "", "absent"); err != nil {
		t.Fatalf("want success for an absent fact, got %v", err)
	}
}

// Some xvec paths report the miss only through the per-key result, without a
// batch error, so that shape has to be handled too.
func TestRemoveFactByIDMissingFactPerKeyOnly(t *testing.T) {
	cause := fmt.Errorf("xvec: DELETE: db: primary key not found: %q: %w", "absent", xvec.ErrNotFound)

	kv := newFakeKV(&fakeCollection{
		deleted: []xvec.WriteResult{{PrimaryKey: "absent", Err: cause}},
	})

	if err := kv.RemoveFactByID(context.Background(), "", "absent"); err != nil {
		t.Fatalf("want success for an absent fact, got %v", err)
	}
}

// A real failure that is not a miss must still be reported as itself: turning
// every error into ErrFactDoesNotExist would hide an IO problem as a 404.
func TestRemoveFactByIDPropagatesRealError(t *testing.T) {
	boom := errors.New("disk on fire")

	kv := newFakeKV(&fakeCollection{
		deleted:   []xvec.WriteResult{{PrimaryKey: "present"}},
		deleteErr: boom,
	})

	err := kv.RemoveFactByID(context.Background(), "", "present")

	if !errors.Is(err, boom) {
		t.Fatalf("want the underlying error, got %v", err)
	}

	if errors.Is(err, apperrors.ErrFactDoesNotExist) {
		t.Fatalf("a real failure must not be reported as a missing fact")
	}
}

// A successful removal stays a no-error call.
func TestRemoveFactByIDPresentFact(t *testing.T) {
	kv := newFakeKV(&fakeCollection{deleted: []xvec.WriteResult{{
		PrimaryKey: "present",
	}}})

	if err := kv.RemoveFactByID(context.Background(), "", "present"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A projected field of the wrong type is a storage error, not a panic and not
// a silently empty value.
func TestGetFactByIDWrongFieldType(t *testing.T) {
	kv := newFakeKV(&fakeCollection{fetched: []*xvec.Document{{
		PrimaryKey: "present",
		Fields: map[string]any{
			keyFieldCreatedAt: 42,
			keyFieldSummary:   "a summary",
			keyFieldFact:      "the full text",
		},
	}}})

	_, err := kv.GetFactByID(context.Background(), "", "present")

	if !errors.Is(err, apperrors.ErrDBConvertError) {
		t.Fatalf("want ErrDBConvertError, got %v", err)
	}
}

// unusableIDs are strings xvec refuses as a primary key: an empty key, bytes that
// are not valid UTF-8, and one longer than its limit. None of them can name a
// stored fact. xvec answers each with invalidArgument, which would surface as a
// 500 for input the caller chose, so the repository checks before it opens the
// collection: the read calls it absent, the delete treats it as nothing to do.
//
// The empty ID never reaches here over HTTP — ogen rejects it while decoding the
// path parameter — but the MCP tools pass `id` through unvalidated, so this is the
// layer that has to answer it.
func unusableIDs() []string {
	return []string{
		"",
		string([]byte{0xff}),
		strings.Repeat("a", xvec.MaxPrimaryKeyBytes+1),
	}
}

func TestGetFactByIDRejectsUnusableID(t *testing.T) {
	for _, id := range unusableIDs() {
		kv := New(testConfig{}, "", testLogger())
		kv.open = func(context.Context) (collection, error) {
			t.Fatalf("opened the collection for the unusable id %q", id)

			return nil, nil
		}

		_, err := kv.GetFactByID(context.Background(), "", id)

		if !errors.Is(err, apperrors.ErrFactDoesNotExist) {
			t.Fatalf("GetFactByID(%q): want ErrFactDoesNotExist, got %v", id, err)
		}
	}
}

func TestRemoveFactByIDTreatsUnusableIDAsNoop(t *testing.T) {
	for _, id := range unusableIDs() {
		kv := New(testConfig{}, "", testLogger())
		kv.open = func(context.Context) (collection, error) {
			t.Fatalf("opened the collection for the unusable id %q", id)

			return nil, nil
		}

		// The same answer an absent fact gets: there is nothing to remove, so
		// the call succeeds. Opening the store would only turn a guarantee into
		// an error from xvec's key validation.
		if err := kv.RemoveFactByID(context.Background(), "", id); err != nil {
			t.Fatalf("RemoveFactByID(%q): want success, got %v", id, err)
		}
	}
}
