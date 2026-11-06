package web

import (
	"context"
	"testing"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

// userStore answers with a fixed user and error, standing in for the real token
// lookup so the controller can be exercised without reading any config.
type userStore struct {
	user user.User
	err  error
}

func (s userStore) GetUserByToken(_ context.Context, _ string) (user.User, error) {
	return s.user, s.err
}

// The constructors differ only in the mode they fix, so a test builds the one it
// needs instead of arranging a config file.
type stubConfig struct{}

func (stubConfig) GetWebHTTPHost() string { return "" }

func controllerWithAuth(t *testing.T, auth Authentication) *Controller {
	t.Helper()

	return NewWithAuth(stubConfig{}, nil, nil, nil, nil, nil, auth, testLogger())
}

func controllerWithoutAuth(t *testing.T, auth Authentication) *Controller {
	t.Helper()

	return NewWithoutAuth(stubConfig{}, nil, nil, nil, nil, nil, auth, testLogger())
}

// With authentication off the header is ignored entirely: the embedded backend
// has no users to resolve against, so a stale token must not break a request.
func TestHandleBearerAuthIgnoredWhenAuthIsOff(t *testing.T) {
	c := controllerWithoutAuth(t, userStore{err: apperrors.ErrUserNotFound})

	ctx, err := c.HandleBearerAuth(context.Background(), server.AddFactOperation, server.BearerAuth{Token: "stale"})
	if err != nil {
		t.Fatalf("HandleBearerAuth: %v", err)
	}

	if got, ok := ctx.Value(userIDKey{}).(string); ok {
		t.Errorf("context carries userID %q although auth is off", got)
	}
}

// A rejected token must not leave a user behind, and must reach the caller as an
// error so the response becomes 401.
func TestHandleBearerAuthDoesNotStoreUserOnError(t *testing.T) {
	c := controllerWithAuth(t, userStore{user: user.User{ID: "u1"}, err: apperrors.ErrUserNotFound})

	ctx, err := c.HandleBearerAuth(context.Background(), server.AddFactOperation, server.BearerAuth{Token: "bad"})
	if err == nil {
		t.Fatal("err = nil, want the store's error")
	}

	if got, ok := ctx.Value(userIDKey{}).(string); ok {
		t.Errorf("context carries userID %q after a rejection", got)
	}
}

// A resolved token must put exactly the ID in the context, since that is what
// every handler reads to pick its collection.
func TestHandleBearerAuthPutsUserID(t *testing.T) {
	c := controllerWithAuth(t, userStore{user: user.User{ID: "u1", Tokens: []string{"tok"}}})

	ctx, err := c.HandleBearerAuth(context.Background(), server.AddFactOperation, server.BearerAuth{Token: "tok"})
	if err != nil {
		t.Fatalf("HandleBearerAuth: %v", err)
	}

	m := &Methods{authRequired: true, logger: testLogger()}

	got, err := m.getUserID(ctx)
	if err != nil {
		t.Fatalf("getUserID: %v", err)
	}

	if got != "u1" {
		t.Errorf("userID = %q, want %q", got, "u1")
	}
}

// The two constructors must differ only in the mode, since that field is what
// gates every handler and what gets baked into the page.
func TestConstructorsSetTheMode(t *testing.T) {
	tests := []struct {
		name string
		c    *Controller
		want bool
	}{
		{name: "with auth", c: controllerWithAuth(t, userStore{}), want: true},
		{name: "without auth", c: controllerWithoutAuth(t, userStore{}), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c.authRequired != tt.want {
				t.Errorf("authRequired = %v, want %v", tt.c.authRequired, tt.want)
			}
		})
	}
}
