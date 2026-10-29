package web

import (
	"context"
	"testing"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

// userStore answers with a fixed user and error, standing in for the real
// token lookup so the controller can be exercised without reading any config.
type userStore struct {
	user user.User
	err  error
}

func (s userStore) GetUserByToken(_ context.Context, _ string) (user.User, error) {
	return s.user, s.err
}

// A rejected token must not leave a user behind. The generated server only keeps
// the context when the handler reports success, but relying on that would make
// the controller's own contract depend on a detail of the generated code.
func TestHandleBearerAuthDoesNotStoreUserOnError(t *testing.T) {
	c := &Controller{
		authentication: userStore{user: user.User{ID: "u1"}, err: apperrors.ErrUserNotFound},
	}

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
	c := &Controller{
		authentication: userStore{user: user.User{ID: "u1", Tokens: []string{"tok"}}},
	}

	ctx, err := c.HandleBearerAuth(context.Background(), server.AddFactOperation, server.BearerAuth{Token: "tok"})
	if err != nil {
		t.Fatalf("HandleBearerAuth: %v", err)
	}

	if got, _ := (&Methods{}).getUserID(ctx); got != "u1" {
		t.Errorf("userID = %q, want %q", got, "u1")
	}
}
