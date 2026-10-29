package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

// stubSecurity is the SecurityHandler the generated server asks for. It lets a
// test pick which answer the handler gives without touching a real token store.
type stubSecurity struct {
	err error
}

func (s stubSecurity) HandleBearerAuth(
	ctx context.Context,
	_ server.OperationName,
	_ server.BearerAuth,
) (context.Context, error) {
	return ctx, s.err
}

// apiFor wires the generated server the way controller.go does and answers with
// whatever the security stage decides, so a request that never reaches a
// handler still exercises the real middleware chain.
func apiFor(t *testing.T, sec server.SecurityHandler) http.Handler {
	t.Helper()

	api, err := server.NewServer(&Methods{}, sec, ServerOptions()...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	return api
}

// A request with no Authorization header is rejected by ogen before our handler
// runs, so the only thing that decides its status is mapDomainError. It must be
// 401: answering 500 would report a client mistake as a server fault.
func TestMissingAuthorizationIsUnauthorized(t *testing.T) {
	api := apiFor(t, stubSecurity{})

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/fact", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeUNAUTHORIZED {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeUNAUTHORIZED)
	}
}

// A token the store does not recognise makes the security handler return
// ErrUserNotFound. ogen wraps it in SecurityError, and the wrapping must not
// cost the specific status.
func TestRejectedTokenIsUnauthorized(t *testing.T) {
	api := apiFor(t, stubSecurity{err: apperrors.ErrUserNotFound})

	r := httptest.NewRequest(http.MethodPost, "/api/v1/fact", nil)
	r.Header.Set("Authorization", "Bearer unknown-token")

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, r)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeUNAUTHORIZED {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeUNAUTHORIZED)
	}
}

// getUserID reads what HandleBearerAuth put in the context. Without a value the
// handlers must refuse rather than fall back to an empty user, which would reach
// the shared base collection.
func TestGetUserIDRequiresContextValue(t *testing.T) {
	m := &Methods{}

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "absent", ctx: context.Background()},
		{name: "empty", ctx: context.WithValue(context.Background(), userIDKey{}, "")},
		{name: "wrong type", ctx: context.WithValue(context.Background(), userIDKey{}, 42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := m.getUserID(tt.ctx); !errors.Is(err, apperrors.ErrUserNotFound) {
				t.Errorf("err = %v, want %v", err, apperrors.ErrUserNotFound)
			}
		})
	}

	if id, err := m.getUserID(context.WithValue(context.Background(), userIDKey{}, "u1")); err != nil || id != "u1" {
		t.Errorf("got (%q, %v), want (%q, nil)", id, err, "u1")
	}
}
