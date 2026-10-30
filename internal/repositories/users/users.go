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

type (
	// backend reports which storage the app is running against. Authentication
	// follows it: Qdrant partitions by user, the embedded index does not.
	backend interface {
		IsQdrantConfigured() bool
	}

	Users struct {
		backend backend
		// configPath is set explicitly by a test, which may not read the real
		// home directory. Empty means "resolve ~/.config/aimem/config.toml".
		configPath string
		// readFile is the seam for tests: the rules forbid touching the filesystem,
		// so a fake supplies the config bytes from memory.
		readFile func(name string) ([]byte, error)
	}
)

func New(b backend) *Users {
	return &Users{backend: b, readFile: os.ReadFile}
}

// NewWithPath builds a repository over an explicit config file path.
func NewWithPath(b backend, configPath string) *Users {
	return &Users{backend: b, configPath: configPath, readFile: os.ReadFile}
}

func (u *Users) path() (string, error) {
	if u.configPath != "" {
		return u.configPath, nil
	}

	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(h, ".config", "aimem", "config.toml"), nil
}

// readUsers parses the [[users]] entries of the config file.
func (u *Users) readUsers() ([]user.User, error) {
	configPath, err := u.path()
	if err != nil {
		return nil, err
	}

	data, err := u.readFile(configPath)
	if err != nil {
		return nil, err
	}

	var ub userBase

	if err := toml.Unmarshal(data, &ub); err != nil {
		return nil, err
	}

	return ub.Users, nil
}

// GetUserByToken resolves a token to the user that owns it.
func (u *Users) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	configured, err := u.readUsers()
	if err != nil {
		return user.User{}, apperrors.ErrUserNotFound
	}

	for _, cur := range configured {
		if slices.Contains(cur.Tokens, token) {
			return cur, nil
		}
	}

	return user.User{}, apperrors.ErrUserNotFound
}

// RequireAuthentication reports whether requests must be authenticated, and
// refuses to start when they must be but no user is configured.
//
// The answer is a property of the storage, not of the config: Qdrant keeps one
// collection per user, so without a token there is no way to know whose memory
// a request means, while the embedded backend stores a single operator's local
// files and ignores the userID entirely. So a `[[users]]` section is required
// for Qdrant and irrelevant for the embedded database — and when it is
// irrelevant, whatever it contains is ignored rather than validated.
//
// The two booleans are deliberately distinct: `true` means "gate the API behind
// a token", `false` means "serve everything openly".
func (u *Users) RequireAuthentication(ctx context.Context) (bool, error) {
	if !u.backend.IsQdrantConfigured() {
		return false, nil
	}

	configured, err := u.readUsers()
	if err != nil || len(configured) == 0 {
		return false, apperrors.ErrNoUsersConfigured
	}

	return true, nil
}
