package users

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

type (
	backend interface {
		RequireAuth() bool
		GetUsersJSON() string
	}

	Users struct {
		backend backend
		list    func() map[string]user.User
	}

	userParsed struct {
		ID     string   `json:"id"`
		Tokens []string `json:"tokens"`
	}
)

func New(b backend) *Users {
	usersJSON := b.GetUsersJSON()
	getUsers := func() map[string]user.User {
		var usersParsed []userParsed

		result := make(map[string]user.User)

		err := json.Unmarshal([]byte(usersJSON), &usersParsed)
		if err != nil {
			return result
		}

		for _, up := range usersParsed {
			id := up.ID
			for _, token := range up.Tokens {
				result[token] = user.User{
					ID:     id,
					Tokens: up.Tokens,
				}
			}
		}

		return result
	}

	return &Users{
		backend: b,
		list:    sync.OnceValue[map[string]user.User](getUsers),
	}
}

// GetUserByToken resolves a token to the user that owns it.
func (u *Users) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	usr, ok := u.list()[token]
	if !ok {
		return user.User{}, apperrors.ErrUserNotFound
	}

	return usr, nil
}

// RequireAuthentication reports whether requests must be authenticated, and
// refuses to start when they must be but no user is configured.
//
// The answer is a property of the storage, not of the config: Qdrant keeps one
// collection per user, so without a token there is no way to know whose memory
// a request means, while the embedded backend stores a single operator's local
// files and ignores the userID entirely. So `AIMEM_USERS_JSON` is required for
// Qdrant and irrelevant for the embedded database — and when it is irrelevant,
// whatever it contains is ignored rather than validated.
//
// The two booleans are deliberately distinct: `true` means "gate the API behind
// a token", `false` means "serve everything openly".
func (u *Users) RequireAuthentication(ctx context.Context) (bool, error) {
	if !u.backend.RequireAuth() {
		return false, nil
	}

	configured := u.list()
	if len(configured) == 0 {
		return false, apperrors.ErrNoUsersConfigured
	}

	return true, nil
}
