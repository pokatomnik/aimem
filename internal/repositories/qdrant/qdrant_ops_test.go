package qdrant

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func TestGetFactsListDoesNotInvertTheScore(t *testing.T) {
	t.Parallel()

	// Qdrant reports cosine similarity directly: 0.9 is close, 0.1 is far. kv
	// reports a distance and flips it with "1 - Score". If this repository
	// applied the same inversion, the 0.9 result would become 0.1 and be filtered
	// out by a 0.5 threshold, while the 0.1 result would become 0.9 and pass.
	repository, _ := newTestRepository(collectionHandlers(func(request recordedRequest) (int, string) {
		if request.Path != pathPointsQuery(testUserID) {
			return http.StatusNotFound, `{"status":{"error":"unexpected path"}}`
		}

		return http.StatusOK, `{
			"result": {
				"points": [
					{"id": "01234567-89ab-cdef-0123-456789abcdef", "score": 0.9, "payload": {"summary": "close", "created_at": "2026-01-01T00:00:00Z"}},
					{"id": "11234567-89ab-cdef-0123-456789abcdef", "score": 0.1, "payload": {"summary": "far", "created_at": "2026-01-01T00:00:00Z"}}
				]
			},
			"status": "ok"
		}`
	}))

	facts, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 5, 0.5)
	if err != nil {
		t.Fatalf("GetFactsList returned error: %v", err)
	}

	if len(facts) != 1 {
		t.Fatalf("got %d facts, want 1: %+v", len(facts), facts)
	}

	if facts[0].Summary != "close" {
		t.Fatalf("kept %q, want the high-score result \"close\"", facts[0].Summary)
	}
}

func TestGetFactsListSendsNamedVectorAndTopK(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
	}))

	if _, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 7, 0.5); err != nil {
		t.Fatalf("GetFactsList returned error: %v", err)
	}

	last := transport.requests[len(transport.requests)-1]

	var body queryPointsRequest
	mustDecodeBody(t, last.Body, &body)

	if body.Using != keyFieldEmbedding {
		t.Fatalf("using = %q, want %q", body.Using, keyFieldEmbedding)
	}

	if body.Limit != 7 {
		t.Fatalf("limit = %d, want 7", body.Limit)
	}

	if len(body.Query) != 3 {
		t.Fatalf("query has %d values, want 3", len(body.Query))
	}
}

func TestGetFactsListRejectsNonPositiveTopK(t *testing.T) {
	t.Parallel()

	repository, _ := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
	}))

	if _, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 0, 0.5); err == nil {
		t.Fatalf("expected an error for topK = 0")
	}
}

func TestGetFactsListReportsConversionErrors(t *testing.T) {
	t.Parallel()

	repository, _ := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{
			"result": {"points": [{"id": "01234567-89ab-cdef-0123-456789abcdef", "score": 0.9, "payload": {"summary": 42, "created_at": "2026-01-01T00:00:00Z"}}]},
			"status": "ok"
		}`
	}))

	_, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 5, 0.5)
	if !errors.Is(err, apperrors.ErrDBConvertError) {
		t.Fatalf("err = %v, want ErrDBConvertError", err)
	}
}

func TestAddFactStoresPayloadAndNamedVector(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(request recordedRequest) (int, string) {
		if request.Path == pathPoints(testUserID) {
			return http.StatusOK, `{"result":[],"status":"ok"}`
		}

		return http.StatusOK, `{"result":{"status":"completed"},"status":"ok"}`
	}))

	err := repository.AddFact(context.Background(), testUserID, []float32{1, 2, 3}, "a summary", "the fact")
	if err != nil {
		t.Fatalf("AddFact returned error: %v", err)
	}

	upsert := transport.requests[len(transport.requests)-1]

	if upsert.Method != http.MethodPut {
		t.Fatalf("method = %s, want PUT", upsert.Method)
	}

	if upsert.Query != "wait=true" {
		t.Fatalf("query = %q, want wait=true so the fact is immediately readable", upsert.Query)
	}

	var body upsertPointsRequest
	mustDecodeBody(t, upsert.Body, &body)

	if len(body.Points) != 1 {
		t.Fatalf("got %d points, want 1", len(body.Points))
	}

	point := body.Points[0]

	if point.ID != newPointID("the fact") {
		t.Fatalf("id = %q, want the digest-derived UUID", point.ID)
	}

	if got := point.Payload[keyFieldSummary]; got != "a summary" {
		t.Fatalf("payload summary = %v, want \"a summary\"", got)
	}

	if got := point.Payload[keyFieldFact]; got != "the fact" {
		t.Fatalf("payload fact = %v, want \"the fact\"", got)
	}

	if _, ok := point.Payload[keyFieldCreatedAt].(string); !ok {
		t.Fatalf("payload created_at = %v, want a string", point.Payload[keyFieldCreatedAt])
	}

	if vector, ok := point.Vector[keyFieldEmbedding]; !ok || len(vector) != 3 {
		t.Fatalf("vector = %v, want %d values under %q", point.Vector, 3, keyFieldEmbedding)
	}
}

func TestAddFactRejectsDuplicate(t *testing.T) {
	t.Parallel()

	// Qdrant's upsert would silently overwrite; kv reports the duplicate.
	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{
			"result": [{"id": "01234567-89ab-cdef-0123-456789abcdef", "payload": {}}],
			"status": "ok"
		}`
	}))

	err := repository.AddFact(context.Background(), testUserID, []float32{1, 2, 3}, "summary", "the fact")
	if !errors.Is(err, apperrors.ErrAlreadyHasFact) {
		t.Fatalf("err = %v, want ErrAlreadyHasFact", err)
	}

	for _, request := range transport.requests {
		if request.Method == http.MethodPut {
			t.Fatalf("AddFact issued an upsert for an existing fact")
		}
	}
}

func TestGetFactByIDReturnsFullView(t *testing.T) {
	t.Parallel()

	const pointID = "01234567-89ab-cdef-0123-456789abcdef"

	repository, _ := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{
			"result": [{"id": "` + pointID + `", "payload": {"summary": "s", "fact": "f", "created_at": "2026-01-01T00:00:00Z"}}],
			"status": "ok"
		}`
	}))

	fact, err := repository.GetFactByID(context.Background(), testUserID, pointID)
	if err != nil {
		t.Fatalf("GetFactByID returned error: %v", err)
	}

	if fact.ID != pointID {
		t.Fatalf("id = %q, want %q", fact.ID, pointID)
	}

	if fact.Summary != "s" || fact.Content != "f" || fact.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("unexpected fact: %+v", fact)
	}
}

func TestGetFactByIDReportsMissingFact(t *testing.T) {
	t.Parallel()

	repository, _ := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":[],"status":"ok"}`
	}))

	_, err := repository.GetFactByID(context.Background(), testUserID, "01234567-89ab-cdef-0123-456789abcdef")
	if !errors.Is(err, apperrors.ErrFactDoesNotExist) {
		t.Fatalf("err = %v, want ErrFactDoesNotExist", err)
	}
}

func TestGetFactByIDAcceptsKVDigest(t *testing.T) {
	t.Parallel()

	// Data carried over from kv arrives as a full SHA-256 digest and must resolve
	// to the same point.
	const digest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":[],"status":"ok"}`
	}))

	_, err := repository.GetFactByID(context.Background(), testUserID, digest)
	if !errors.Is(err, apperrors.ErrFactDoesNotExist) {
		t.Fatalf("err = %v, want ErrFactDoesNotExist", err)
	}

	var body retrievePointsRequest
	mustDecodeBody(t, transport.requests[len(transport.requests)-1].Body, &body)

	if len(body.IDs) != 1 || body.IDs[0] != "01234567-89ab-cdef-0123-456789abcdef" {
		t.Fatalf("ids = %v, want the mapped UUID", body.IDs)
	}
}

func TestRemoveFactByIDDeletesExistingPoint(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(request recordedRequest) (int, string) {
		if request.Path == pathPoints(testUserID) {
			return http.StatusOK, `{"result":[{"id":"01234567-89ab-cdef-0123-456789abcdef","payload":{}}],"status":"ok"}`
		}

		return http.StatusOK, `{"result":{"status":"completed"},"status":"ok"}`
	}))

	err := repository.RemoveFactByID(context.Background(), testUserID, "01234567-89ab-cdef-0123-456789abcdef")
	if err != nil {
		t.Fatalf("RemoveFactByID returned error: %v", err)
	}

	last := transport.requests[len(transport.requests)-1]

	if last.Path != pathPointsDelete(testUserID) {
		t.Fatalf("path = %q, want %q", last.Path, pathPointsDelete(testUserID))
	}

	if last.Query != "wait=true" {
		t.Fatalf("query = %q, want wait=true", last.Query)
	}

	var body deletePointsRequest
	mustDecodeBody(t, last.Body, &body)

	if len(body.Points) != 1 || body.Points[0] != "01234567-89ab-cdef-0123-456789abcdef" {
		t.Fatalf("points = %v, want the target ID", body.Points)
	}
}

// Deleting a fact that is not stored succeeds: DELETE is idempotent, so a
// client can retry it and get the same answer as the first call. The embedded
// side reports an absent key as a batch error and a per-key error, and both mean
// "nothing to delete"; here the delete endpoint simply succeeds.
func TestRemoveFactByIDReportsMissingFact(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":[],"status":"ok"}`
	}))

	if err := repository.RemoveFactByID(context.Background(), testUserID, "01234567-89ab-cdef-0123-456789abcdef"); err != nil {
		t.Fatalf("RemoveFactByID returned error: %v", err)
	}

	last := transport.requests[len(transport.requests)-1]

	if last.Path != pathPointsDelete(testUserID) {
		t.Fatalf("path = %q, want %q", last.Path, pathPointsDelete(testUserID))
	}
}

// An ID Qdrant cannot store is a fact that does not exist, not a conversion
// error. Qdrant accepts a point ID only as a UUID or a u64, so a string like
// "not-an-id" cannot name a point in any collection; reporting ErrDBConvertError
// would turn a caller's mistake — or an ID minted when the embedded backend was
// selected — into 500 DBCONVERTERROR, while the embedded backend answers 404.
//
// The check runs before the collection bootstrap on purpose: there is nothing to
// provision for an ID that cannot be stored, so the server is never asked.
func TestGetFactByIDTreatsUnusableIDAsMissing(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":[],"status":"ok"}`
	}))

	_, err := repository.GetFactByID(context.Background(), testUserID, "not-an-id")
	if !errors.Is(err, apperrors.ErrFactDoesNotExist) {
		t.Fatalf("err = %v, want ErrFactDoesNotExist", err)
	}

	if len(transport.requests) != 0 {
		t.Fatalf("asked the server about an unstorable ID: %+v", transport.requests)
	}
}

// An ID that cannot name a point still deletes successfully: there is nothing
// behind it, so the request is a no-op. The server is not asked and no
// collection is provisioned, because neither could hold such a fact.
func TestRemoveFactByIDTreatsUnusableIDAsNoop(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":[],"status":"ok"}`
	}))

	if err := repository.RemoveFactByID(context.Background(), testUserID, "not-an-id"); err != nil {
		t.Fatalf("err = %v, want success", err)
	}

	if len(transport.requests) != 0 {
		t.Fatalf("asked the server about an unstorable ID: %+v", transport.requests)
	}
}

func TestFactsAreScopedPerUser(t *testing.T) {
	t.Parallel()

	// A point ID is derived from the fact's text alone, so it carries no user. Two
	// users storing the same text therefore address the same point ID, and a shared
	// collection would answer the second AddFact with ErrAlreadyHasFact. The
	// per-user collection is what keeps them apart.
	const (
		userA = "a"
		userB = "b"
		text  = "the same fact"
	)

	// stored records which collection received the upsert, so the second user's
	// existence check reads its own collection instead of the first user's.
	stored := map[string]bool{}

	repository, transport := newTestRepository(func(request recordedRequest) (int, string) {
		if strings.HasSuffix(request.Path, "/exists") {
			return http.StatusOK, `{"result":{"exists":true},"status":"ok"}`
		}

		if request.Path == pathPoints(userA) || request.Path == pathPoints(userB) {
			if request.Method == http.MethodPut {
				stored[request.Path] = true

				return http.StatusOK, `{"result":{"status":"completed"},"status":"ok"}`
			}

			if stored[request.Path] {
				return http.StatusOK, `{"result":[{"id":"` + newPointID(text) + `","payload":{}}],"status":"ok"}`
			}

			return http.StatusOK, `{"result":[],"status":"ok"}`
		}

		return http.StatusNotFound, `{"status":{"error":"unexpected path"}}`
	})

	if err := repository.AddFact(context.Background(), userA, []float32{1, 0, 0}, "s", text); err != nil {
		t.Fatalf("AddFact for the first user returned error: %v", err)
	}

	if err := repository.AddFact(context.Background(), userB, []float32{1, 0, 0}, "s", text); err != nil {
		t.Fatalf("AddFact for the second user returned error: %v", err)
	}

	wantPaths := []string{pathPoints(userA), pathPoints(userB)}
	upserts := make([]string, 0, len(wantPaths))

	for _, request := range transport.requests {
		if request.Method == http.MethodPut {
			upserts = append(upserts, request.Path)
		}
	}

	if len(upserts) != len(wantPaths) {
		t.Fatalf("upserts = %v, want one per user", upserts)
	}

	for i, want := range wantPaths {
		if upserts[i] != want {
			t.Fatalf("upsert %d went to %q, want %q", i, upserts[i], want)
		}
	}
}

func TestCollectionIsCreatedWhenMissing(t *testing.T) {
	t.Parallel()

	repository, transport := newTestRepository(func(request recordedRequest) (int, string) {
		switch request.Path {
		case pathCollectionExists(testUserID):
			return http.StatusOK, `{"result":{"exists":false},"status":"ok"}`
		case pathCollection(testUserID):
			return http.StatusOK, `{"result":true,"status":"ok"}`
		}

		return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
	})

	if _, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 5, 0.5); err != nil {
		t.Fatalf("GetFactsList returned error: %v", err)
	}

	var created createCollectionRequest
	found := false

	for _, request := range transport.requests {
		if request.Path != pathCollection(testUserID) || request.Method != http.MethodPut {
			continue
		}

		found = true
		mustDecodeBody(t, request.Body, &created)
	}

	if !found {
		t.Fatalf("collection was not created")
	}

	vector, ok := created.Vectors[keyFieldEmbedding]
	if !ok {
		t.Fatalf("collection has vectors %v, want %q", created.Vectors, keyFieldEmbedding)
	}

	if vector.Distance != distanceCosine {
		t.Fatalf("distance = %q, want %q", vector.Distance, distanceCosine)
	}

	if vector.Size != 3 {
		t.Fatalf("size = %d, want the configured 3", vector.Size)
	}
}

func TestCollectionCreationToleratesConcurrentCreate(t *testing.T) {
	t.Parallel()

	repository, _ := newTestRepository(func(request recordedRequest) (int, string) {
		switch request.Path {
		case pathCollectionExists(testUserID):
			return http.StatusOK, `{"result":{"exists":false},"status":"ok"}`
		case pathCollection(testUserID):
			return http.StatusConflict, `{"status":{"error":"already exists"}}`
		}

		return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
	})

	if _, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 5, 0.5); err != nil {
		t.Fatalf("a concurrent create should not fail the operation: %v", err)
	}
}

func TestAPIKeyIsSentWhenConfigured(t *testing.T) {
	t.Parallel()

	transport := &fakeTransport{handle: collectionHandlers(func(recordedRequest) (int, string) {
		return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
	})}

	repository := New(fakeConfig{
		host:       "http://qdrant.test",
		apiKey:     "secret",
		dimensions: 3,
	}, &http.Client{Transport: transport})

	if _, err := repository.GetFactsList(context.Background(), testUserID, []float32{1, 0, 0}, 5, 0.5); err != nil {
		t.Fatalf("GetFactsList returned error: %v", err)
	}

	for _, request := range transport.requests {
		if got := request.Header.Get(headerAPIKey); got != "secret" {
			t.Fatalf("api-key header = %q, want \"secret\"", got)
		}
	}
}
