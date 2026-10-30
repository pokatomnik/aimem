package tools

import (
	"context"
	"errors"
	"net/http"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp/mcpdeps"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
	factentity "github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

// required is the authenticator most tests use: every call must carry a valid
// token, because that is the production wiring for Qdrant.
func required(f *fakeDeps) mcpdeps.Authenticator {
	return NewTokenAuthenticator(f)
}

// anonymous stands in for the embedded deployment: no credentials are read at all.
func anonymous() mcpdeps.Authenticator {
	return NewAnonymousAuthenticator()
}

// fakeDeps records what reached it so a test can prove the user resolved from the
// token is the one the store operation received.
type fakeDeps struct {
	recalled string
	readID   string
	userID   string

	// authCalls counts token lookups, so a test can tell "authenticate was never
	// called" from "it was called and happened to answer with an empty ID".
	authCalls int

	authErr error
}

func (f *fakeDeps) Remember(ctx context.Context, userID, s string) (string, error) {
	f.userID = userID
	return "summary of " + s, nil
}

func (f *fakeDeps) Recall(ctx context.Context, userID, q string) ([]factentity.FactListView, error) {
	f.recalled = q
	f.userID = userID
	return []factentity.FactListView{
		{ID: "id-1", CreatedAt: "2026-01-01", Summary: "short summary one"},
		{ID: "id-2", CreatedAt: "2026-01-02", Summary: "short summary two"},
	}, nil
}

func (f *fakeDeps) GetFact(ctx context.Context, userID, id string) (factentity.FactFullView, error) {
	f.readID = id
	f.userID = userID
	return factentity.FactFullView{
		ID:        id,
		CreatedAt: "2026-01-01",
		Summary:   "short",
		Content:   "the full stored text",
	}, nil
}

func (f *fakeDeps) RemoveFactByID(ctx context.Context, userID, id string) error {
	f.userID = userID
	return nil
}

func (f *fakeDeps) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	f.authCalls++

	if f.authErr != nil {
		return user.User{}, f.authErr
	}

	return user.User{ID: "u-" + token, Tokens: []string{token}}, nil
}

// reqWithToken builds the request a tools/call carries over the streamable HTTP
// transport, where the credentials live in Extra.Header rather than in the tool
// arguments.
func reqWithToken(value string) *sdk.CallToolRequest {
	h := http.Header{}
	if value != "" {
		h.Set("Authorization", value)
	}

	return &sdk.CallToolRequest{Extra: &sdk.RequestExtra{Header: h}}
}

func TestRecallReturnsSummariesOnly(t *testing.T) {
	f := &fakeDeps{}

	_, handler, err := ToolRecall(required(f), f)
	if err != nil {
		t.Fatal(err)
	}

	_, out, err := handler(context.Background(), reqWithToken("Bearer tok"), recallInput{Query: "anything"})
	if err != nil {
		t.Fatalf("recall failed: %v", err)
	}

	if len(out.Facts) != 2 {
		t.Fatalf("got %d facts, want 2", len(out.Facts))
	}

	if out.Facts[0].Summary != "short summary one" {
		t.Errorf("summary = %q, want %q", out.Facts[0].Summary, "short summary one")
	}

	if f.recalled != "anything" {
		t.Errorf("query not forwarded, got %q", f.recalled)
	}
}

func TestReadFactReturnsFullText(t *testing.T) {
	f := &fakeDeps{}

	_, handler, err := ToolGetFact(required(f), f)
	if err != nil {
		t.Fatal(err)
	}

	_, out, err := handler(context.Background(), reqWithToken("Bearer tok"), getFactInput{ID: "id-1"})
	if err != nil {
		t.Fatalf("read_fact failed: %v", err)
	}

	if out.Fact != "the full stored text" {
		t.Errorf("fact = %q, want the full text", out.Fact)
	}

	if f.readID != "id-1" {
		t.Errorf("id not forwarded, got %q", f.readID)
	}
}

// The user resolved from the token must be the one the store receives, otherwise
// every caller would read and write the same memory.
func TestUserResolvedFromTokenReachesStore(t *testing.T) {
	tests := []struct {
		name string
		run  func(f *fakeDeps) error
	}{
		{
			name: "remember",
			run: func(f *fakeDeps) error {
				_, h, err := ToolRemember(required(f), f)
				if err != nil {
					return err
				}
				_, _, err = h(context.Background(), reqWithToken("Bearer alice"), rememberInput{Fact: "x"})
				return err
			},
		},
		{
			name: "recall",
			run: func(f *fakeDeps) error {
				_, h, err := ToolRecall(required(f), f)
				if err != nil {
					return err
				}
				_, _, err = h(context.Background(), reqWithToken("Bearer alice"), recallInput{Query: "x"})
				return err
			},
		},
		{
			name: "read_fact",
			run: func(f *fakeDeps) error {
				_, h, err := ToolGetFact(required(f), f)
				if err != nil {
					return err
				}
				_, _, err = h(context.Background(), reqWithToken("Bearer alice"), getFactInput{ID: "id-1"})
				return err
			},
		},
		{
			name: "remove_fact_by_id",
			run: func(f *fakeDeps) error {
				_, h, err := ToolRemoveFact(required(f), f)
				if err != nil {
					return err
				}
				_, _, err = h(context.Background(), reqWithToken("Bearer alice"), removeFactInput{ID: "id-1"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeDeps{}

			if err := tt.run(f); err != nil {
				t.Fatalf("tool failed: %v", err)
			}

			if f.userID != "u-alice" {
				t.Errorf("store got userID %q, want %q", f.userID, "u-alice")
			}
		})
	}
}

// No credentials, or credentials the store rejects, must be an ordinary error
// answer. In particular a request without Extra must not panic: a panic leaves
// the call unanswered instead of telling the caller it was refused.
func TestToolRefusesWithoutValidToken(t *testing.T) {
	tests := []struct {
		name  string
		req   *sdk.CallToolRequest
		store *fakeDeps
	}{
		{
			name:  "no extra at all",
			req:   &sdk.CallToolRequest{},
			store: &fakeDeps{},
		},
		{
			name:  "no header",
			req:   reqWithToken(""),
			store: &fakeDeps{},
		},
		{
			name:  "header without the bearer scheme",
			req:   reqWithToken("tok"),
			store: &fakeDeps{},
		},
		{
			name:  "unknown token",
			req:   reqWithToken("Bearer tok"),
			store: &fakeDeps{authErr: apperrors.ErrUserNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, handler, err := ToolRemember(required(tt.store), tt.store)
			if err != nil {
				t.Fatal(err)
			}

			_, _, err = handler(context.Background(), tt.req, rememberInput{Fact: "x"})
			if !errors.Is(err, apperrors.ErrUserNotFound) {
				t.Errorf("err = %v, want %v", err, apperrors.ErrUserNotFound)
			}

			if tt.store.userID != "" {
				t.Errorf("store was reached with userID %q despite a refusal", tt.store.userID)
			}
		})
	}
}

// With authentication off the tools must run anonymously: the embedded backend
// has no users to resolve, so demanding a token there would break a deployment
// that never configured one.
func TestToolRunsAnonymouslyWhenAuthIsOff(t *testing.T) {
	f := &fakeDeps{}

	_, handler, err := ToolRemember(anonymous(), f)
	if err != nil {
		t.Fatal(err)
	}

	// No headers at all: there is nothing to read and nothing to check.
	_, out, err := handler(context.Background(), &sdk.CallToolRequest{}, rememberInput{Fact: "x"})
	if err != nil {
		t.Fatalf("remember failed: %v", err)
	}

	if f.authCalls != 0 {
		t.Errorf("GetUserByToken called %d times, want 0: credentials must not be read when none are needed", f.authCalls)
	}

	if out.Summary != "summary of x" {
		t.Errorf("summary = %q, want the store's answer", out.Summary)
	}

	if f.userID != "" {
		t.Errorf("userID = %q, want the empty one the embedded backend ignores", f.userID)
	}
}

// The token authenticator is injected, not chosen per call: a deployment that
// requires tokens never falls back to anonymous access, whatever the request looks
// like.
func TestTokenAuthenticatorRejectsEveryUnusableCredential(t *testing.T) {
	tests := []struct {
		name string
		req  *sdk.CallToolRequest
	}{
		{name: "no request", req: nil},
		{name: "no extra", req: &sdk.CallToolRequest{}},
		{name: "no header", req: reqWithToken("")},
		{name: "wrong scheme", req: reqWithToken("tok")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeDeps{}

			if _, err := NewTokenAuthenticator(f).UserID(context.Background(), tt.req); !errors.Is(err, apperrors.ErrUserNotFound) {
				t.Errorf("err = %v, want %v", err, apperrors.ErrUserNotFound)
			}

			if f.authCalls != 0 {
				t.Errorf("store was consulted %d times, want 0", f.authCalls)
			}
		})
	}
}

// The anonymous authenticator reads nothing at all: it is the whole implementation
// of "this deployment has no users".
func TestAnonymousAuthenticatorReadsNothing(t *testing.T) {
	f := &fakeDeps{}

	id, err := NewAnonymousAuthenticator().UserID(context.Background(), reqWithToken("Bearer ignored"))
	if err != nil {
		t.Fatalf("UserID: %v", err)
	}

	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}

	if f.authCalls != 0 {
		t.Errorf("store was consulted %d times, want 0", f.authCalls)
	}
}
