package qdrant

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// The store's hard rule is that records are immutable: a point is written whole
// and never patched. That rules out the payload and vector update endpoints,
// which change a stored point in place. A fact that had to change would be
// removed and written again, so those endpoints must stay unreachable from this
// package no matter which operation runs.
//
// The list is the full set of in-place mutation routes in the pinned Qdrant API
// that a partially written record could come from.
var inPlaceMutationPaths = []string{
	pathPoints + "/payload",
	pathPoints + "/payload/delete",
	pathPoints + "/payload/clear",
	pathPoints + "/vectors",
	pathPoints + "/vectors/delete",
	pathPoints + "/batch",
}

func TestNoOperationPatchesAStoredPoint(t *testing.T) {
	t.Parallel()

	const id = "01234567-89ab-cdef-0123-456789abcdef"

	// Every operation runs against a fake that answers the collection bootstrap, a
	// lookup and a successful write, so all four reach their write or read call
	// rather than failing early. The lookup answers "exists" only for the ID the
	// read and delete operations use, so AddFact still reaches its insert.
	respond := collectionHandlers(func(request recordedRequest) (int, string) {
		switch request.Path {
		case pathPointsQuery:
			return http.StatusOK, `{"result":{"points":[]},"status":"ok"}`
		case pathPoints:
			if request.Method == http.MethodPut {
				return http.StatusOK, `{"result":{"status":"completed"},"status":"ok"}`
			}

			if strings.Contains(string(request.Body), id) {
				return http.StatusOK, `{"result":[{"id":"` + id + `","payload":{"summary":"s","fact":"f","created_at":"2026-01-01T00:00:00Z"}}],"status":"ok"}`
			}

			return http.StatusOK, `{"result":[],"status":"ok"}`
		}

		return http.StatusOK, `{"result":{"status":"completed"},"status":"ok"}`
	})

	operations := map[string]func(context.Context, *Qdrant) error{
		"AddFact": func(ctx context.Context, q *Qdrant) error {
			return q.AddFact(ctx, []float32{1, 0, 0}, "s", "f")
		},
		"GetFactByID": func(ctx context.Context, q *Qdrant) error {
			_, err := q.GetFactByID(ctx, id)

			return err
		},
		"GetFactsList": func(ctx context.Context, q *Qdrant) error {
			_, err := q.GetFactsList(ctx, []float32{1, 0, 0}, 5, 0.5)

			return err
		},
		"RemoveFactByID": func(ctx context.Context, q *Qdrant) error {
			return q.RemoveFactByID(ctx, id)
		},
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repository, transport := newTestRepository(respond)

			if err := operation(context.Background(), repository); err != nil {
				t.Fatalf("%s returned error: %v", name, err)
			}

			if len(transport.requests) == 0 {
				t.Fatalf("%s issued no request", name)
			}

			for _, request := range transport.requests {
				for _, forbidden := range inPlaceMutationPaths {
					if request.Path == forbidden {
						t.Fatalf("%s reached %s %s, which rewrites a stored point in place",
							name, request.Method, request.Path)
					}
				}
			}
		})
	}
}
