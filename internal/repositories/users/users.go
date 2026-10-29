package users

import (
	"context"
	"os"
	"path"
	"slices"

	"github.com/pelletier/go-toml/v2"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

type Users struct {
}

func New() *Users {
	return &Users{}
}

func (u *Users) getUsers() ([]user.User, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return []user.User{}, err
	}

	path := path.Join(h, ".config", "aimem", "config.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		return []user.User{}, err
	}

	var ub userBase

	err = toml.Unmarshal(data, &ub)
	if err != nil {
		return []user.User{}, err
	}

	return ub.Users, nil
}

func (u *Users) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	users, err := u.getUsers()
	if err != nil {
		return user.User{}, apperrors.ErrUserNotFound
	}

	for _, u := range users {
		if slices.Contains(u.Tokens, token) {
			return u, nil
		}
	}

	return user.User{}, apperrors.ErrUserNotFound
}
