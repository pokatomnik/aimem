package users

import "github.com/pokatomnik/gomem/internal/entities/user"

type userBase struct {
	Users []user.User `toml:"users"`
}
