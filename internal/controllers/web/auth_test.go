package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// apiFor wires the generated server the way controller.go does.
func apiFor(t *testing.T, sec server.SecurityHandler, authRequired bool, m *Methods) http.Handler {
	t.Helper()

	if m == nil {
		m = &Methods{}
	}
	m.authRequired = authRequired
	if m.logger == nil {
		m.logger = testLogger()
	}

	api, err := server.NewServer(m, sec, ServerOptions(testLogger())...)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	return api
}

// postFact issues a well-formed add-fact request, so the failure under test is
// authentication and not body validation.
func postFact(h http.Handler, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/fact",
		strings.NewReader(`{"value":{"fact":"x"}}`))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)

	return rec
}

// With authentication required and no token, the handler refuses: there is no
// user whose memory could be read or written. 401, not 500 — nothing broke.
func TestMissingTokenIsUnauthorized(t *testing.T) {
	// The security stage accepts the request (the header is optional at the
	// protocol level), so the handler is what must refuse.
	api := apiFor(t, stubSecurity{}, true, nil)

	rec := postFact(api, "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeUNAUTHORIZED {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeUNAUTHORIZED)
	}
}

// A token the store does not recognise makes the security handler return
// ErrUserNotFound. ogen wraps it in SecurityError, and the wrapping must not cost
// the specific status.
func TestRejectedTokenIsUnauthorized(t *testing.T) {
	api := apiFor(t, stubSecurity{err: apperrors.ErrUserNotFound}, true, nil)

	rec := postFact(api, "unknown-token")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body := decodeAppError(t, rec)
	if body.Error != server.ErrorCodeUNAUTHORIZED {
		t.Errorf("error = %q, want %q", body.Error, server.ErrorCodeUNAUTHORIZED)
	}
}

// The embedded backend has nothing to partition, so a request with no token must
// go through untouched — the same anonymous behaviour it had before users existed.
func TestNoTokenIsAcceptedWhenAuthIsOff(t *testing.T) {
	var gotUserID string

	m := &Methods{
		remember: RememberFunc(func(_ context.Context, userID, _ string) (string, error) {
			gotUserID = userID
			return "summary", nil
		}),
	}

	api := apiFor(t, stubSecurity{}, false, m)

	rec := postFact(api, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if gotUserID != "" {
		t.Errorf("userID = %q, want the empty one the embedded backend ignores", gotUserID)
	}
}

// getUserID is the single gate every handler goes through, so its two modes are
// worth pinning directly.
func TestGetUserID(t *testing.T) {
	tests := []struct {
		name         string
		authRequired bool
		ctx          context.Context
		wantID       string
		wantError    bool
	}{
		{
			name:         "auth off, no user",
			authRequired: false,
			ctx:          context.Background(),
		},
		{
			// The context is never consulted when auth is off: there is nothing to look
			// up, and a stale value must not leak into a store that ignores it anyway.
			name:         "auth off never reads the user",
			authRequired: false,
			ctx:          context.WithValue(context.Background(), userIDKey{}, "u1"),
			wantID:       "",
		},
		{
			name:         "auth on, no user",
			authRequired: true,
			ctx:          context.Background(),
			wantError:    true,
		},
		{
			name:         "auth on, empty user",
			authRequired: true,
			ctx:          context.WithValue(context.Background(), userIDKey{}, ""),
			wantError:    true,
		},
		{
			name:         "auth on, wrong type",
			authRequired: true,
			ctx:          context.WithValue(context.Background(), userIDKey{}, 42),
			wantError:    true,
		},
		{
			name:         "auth on, resolved user",
			authRequired: true,
			ctx:          context.WithValue(context.Background(), userIDKey{}, "u1"),
			wantID:       "u1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Methods{authRequired: tt.authRequired, logger: testLogger()}

			id, err := m.getUserID(tt.ctx)

			if tt.wantError {
				if !errors.Is(err, apperrors.ErrUserNotFound) {
					t.Fatalf("err = %v, want %v", err, apperrors.ErrUserNotFound)
				}
				return
			}

			if err != nil {
				t.Fatalf("getUserID: %v", err)
			}

			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
		})
	}
}

// RememberFunc adapts a plain function to the Remember interface, so a test can
// observe what reached the use case without a full fake struct.
type RememberFunc func(ctx context.Context, userID, fact string) (string, error)

func (f RememberFunc) Remember(ctx context.Context, userID, fact string) (string, error) {
	return f(ctx, userID, fact)
}

// countingContext records whether the user ID was looked up at all, so a test can
// tell "the value was never consulted" from "it was consulted and came back
// empty". Only a real Value call counts: the point is that the mode is checked
// before the lookup, not after.
type countingContext struct {
	context.Context
	reads int
}

func (c *countingContext) Value(key any) any {
	if _, ok := key.(userIDKey); ok {
		c.reads++
	}

	return c.Context.Value(key)
}

// With authentication off the user ID must not be fetched at all: the embedded
// backend ignores it, so reading it would only give a stale value a chance to
// travel further than it should.
func TestGetUserIDDoesNotReadContextWhenAuthIsOff(t *testing.T) {
	ctx := &countingContext{Context: context.WithValue(context.Background(), userIDKey{}, "u1")}

	m := &Methods{authRequired: false, logger: testLogger()}

	id, err := m.getUserID(ctx)

	if err != nil {
		t.Fatalf("getUserID: %v", err)
	}

	if id != "" {
		t.Errorf("id = %q, want empty: auth off must not surface a stored user", id)
	}

	if ctx.reads != 0 {
		t.Errorf("context read %d times, want 0: the mode must be checked before the lookup", ctx.reads)
	}
}

// With authentication on the lookup has to happen, or every request would be
// refused for the wrong reason.
func TestGetUserIDReadsContextWhenAuthIsOn(t *testing.T) {
	ctx := &countingContext{Context: context.WithValue(context.Background(), userIDKey{}, "u1")}

	m := &Methods{authRequired: true, logger: testLogger()}

	if _, err := m.getUserID(ctx); err != nil {
		t.Fatalf("getUserID: %v", err)
	}

	if ctx.reads != 1 {
		t.Errorf("context read %d times, want 1", ctx.reads)
	}
}
