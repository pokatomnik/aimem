package users

import (
	"context"
	"errors"
	"testing"

	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

// fakeBackend stands in for the config: only the backend choice matters here.
type fakeBackend struct {
	qdrant bool
}

func (f fakeBackend) IsQdrantConfigured() bool { return f.qdrant }

// configFrom feeds readUsers from memory. The rules forbid touching the
// filesystem from a test, so this replaces os.ReadFile rather than naming a path.
func configFrom(contents string) func(string) ([]byte, error) {
	return func(string) ([]byte, error) {
		return []byte(contents), nil
	}
}

func usersFrom(contents string) *Users {
	return &Users{
		backend:    fakeBackend{qdrant: true},
		configPath: "config.toml",
		readFile:   configFrom(contents),
	}
}

const twoUsers = `
[[users]]
id = "alice"
tokens = ["tok-a"]

[[users]]
id = "bob"
tokens = ["tok-b1", "tok-b2"]
`

func TestGetUserByToken(t *testing.T) {
	u := usersFrom(twoUsers)

	tests := []struct {
		name      string
		token     string
		wantID    string
		wantError bool
	}{
		{name: "first user", token: "tok-a", wantID: "alice"},
		{name: "second user", token: "tok-b1", wantID: "bob"},
		{name: "second token of a user", token: "tok-b2", wantID: "bob"},
		{name: "unknown token", token: "nope", wantError: true},
		{name: "empty token", token: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := u.GetUserByToken(context.Background(), tt.token)

			if tt.wantError {
				if !errors.Is(err, apperrors.ErrUserNotFound) {
					t.Fatalf("err = %v, want %v", err, apperrors.ErrUserNotFound)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetUserByToken: %v", err)
			}

			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}

// A user is exactly an identifier plus its tokens; the config is the only place
// they are declared, so an unparsable file must not be read as "no users".
func TestGetUserByTokenBrokenConfig(t *testing.T) {
	u := usersFrom("this is not toml {{{")

	if _, err := u.GetUserByToken(context.Background(), "tok-a"); !errors.Is(err, apperrors.ErrUserNotFound) {
		t.Errorf("err = %v, want %v", err, apperrors.ErrUserNotFound)
	}
}

// The embedded backend is one operator's local files and ignores userID, so a
// token is not needed there no matter what the file says.
func TestRequireAuthenticationEmbedded(t *testing.T) {
	u := &Users{
		backend:    fakeBackend{qdrant: false},
		configPath: "config.toml",
		readFile:   configFrom(""),
	}

	required, err := u.RequireAuthentication(context.Background())
	if err != nil {
		t.Fatalf("RequireAuthentication: %v", err)
	}

	if required {
		t.Error("auth required with the embedded backend")
	}
}

// A token in the config is ignored by the embedded backend rather than honoured:
// there is nothing to partition, so gating the API would only break a working
// single-user setup.
func TestRequireAuthenticationEmbeddedIgnoresConfiguredUsers(t *testing.T) {
	u := &Users{
		backend:    fakeBackend{qdrant: false},
		configPath: "config.toml",
		readFile:   configFrom(twoUsers),
	}

	required, err := u.RequireAuthentication(context.Background())
	if err != nil {
		t.Fatalf("RequireAuthentication: %v", err)
	}

	if required {
		t.Error("auth required although the backend is the embedded one")
	}
}

func TestRequireAuthenticationQdrant(t *testing.T) {
	tests := []struct {
		name      string
		contents  string
		want      bool
		wantError bool
	}{
		{name: "users configured", contents: twoUsers, want: true},
		{name: "no [[users]] section", contents: "topK = 5\n", wantError: true},
		{name: "empty file", contents: "", wantError: true},
		{name: "unparsable file", contents: "{{{{", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usersFrom(tt.contents)

			got, err := u.RequireAuthentication(context.Background())

			if tt.wantError {
				if !errors.Is(err, apperrors.ErrNoUsersConfigured) {
					t.Fatalf("err = %v, want %v", err, apperrors.ErrNoUsersConfigured)
				}
				if got {
					t.Error("required = true alongside an error")
				}
				return
			}

			if err != nil {
				t.Fatalf("RequireAuthentication: %v", err)
			}

			if got != tt.want {
				t.Errorf("required = %v, want %v", got, tt.want)
			}
		})
	}
}

// A missing file is the same situation as an empty one when Qdrant is selected:
// the app cannot start without anyone to attribute memory to.
func TestRequireAuthenticationQdrantMissingFile(t *testing.T) {
	u := &Users{
		backend:    fakeBackend{qdrant: true},
		configPath: "config.toml",
		readFile: func(string) ([]byte, error) {
			return nil, errors.New("no such file")
		},
	}

	if _, err := u.RequireAuthentication(context.Background()); !errors.Is(err, apperrors.ErrNoUsersConfigured) {
		t.Errorf("err = %v, want %v", err, apperrors.ErrNoUsersConfigured)
	}
}

// The identifier becomes part of a Qdrant collection name, and Qdrant accepts only
// ASCII letters, digits, "_" and "-" there. Documented as a requirement rather
// than enforced, so this pins that a valid one survives parsing untouched.
func TestUserIdentifierRoundTrips(t *testing.T) {
	u := usersFrom("[[users]]\nid = \"team-1_a\"\ntokens = [\"tok\"]\n")

	got, err := u.GetUserByToken(context.Background(), "tok")
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}

	want := user.User{ID: "team-1_a", Tokens: []string{"tok"}}
	if got.ID != want.ID || len(got.Tokens) != 1 || got.Tokens[0] != want.Tokens[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
