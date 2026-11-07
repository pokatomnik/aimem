package mcp

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/entities/user"
)

// fakeDeps satisfies every dependency the server asks for. The tool behaviour
// itself is covered in the tools package, where the input types are visible; here
// it only has to be non-nil. userID records what the last call received, so the
// handler tests can prove the token reached the store.
type fakeDeps struct {
	userID string
}

func (f *fakeDeps) Remember(ctx context.Context, userID, s string) (string, error) {
	f.userID = userID

	return "summary of " + s, nil
}

func (f *fakeDeps) Recall(ctx context.Context, userID, q string) ([]fact.FactListView, error) {
	f.userID = userID

	return []fact.FactListView{
		{ID: "id-1", CreatedAt: "2026-01-01", Summary: "short summary one"},
	}, nil
}

func (f *fakeDeps) GetFact(ctx context.Context, userID, id string) (fact.FactFullView, error) {
	f.userID = userID

	return fact.FactFullView{ID: id}, nil
}

func (f *fakeDeps) RemoveFactByID(ctx context.Context, userID, id string) error {
	f.userID = userID

	return nil
}

// fakeAuth is the token store the authenticated constructor receives. Only the
// resolved ID matters here; the token parsing is covered in the tools package.
type fakeAuth struct{}

func (fakeAuth) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	return user.User{ID: "u-" + token}, nil
}

// connect builds a server through one of the two constructors, so every test goes
// through the same wiring production uses.
type buildFunc func(MCPDeps) (*sdk.Server, error)

func connect(t *testing.T, build buildFunc, deps MCPDeps) (*sdk.ClientSession, context.Context) {
	t.Helper()

	ctx := context.Background()

	server, err := build(deps)
	if err != nil {
		t.Fatal(err)
	}

	serverT, clientT := sdk.NewInMemoryTransports()

	go func() { _ = server.Run(ctx, serverT) }()

	client := sdk.NewClient(&sdk.Implementation{Name: "test"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { cs.Close() })

	return cs, ctx
}

// withAuth is the production wiring for Qdrant: every call needs a token. The
// in-memory transport carries no headers at all, so this exercises the refusing
// side of the answer.
func withAuth(deps MCPDeps) (*sdk.Server, error) {
	srv, err := NewWithAuth(deps, fakeAuth{}, testLogger())
	if err != nil {
		return nil, err
	}

	return srv.server, nil
}

// withoutAuth is the production wiring for the embedded backend.
func withoutAuth(deps MCPDeps) (*sdk.Server, error) {
	srv, err := NewWithoutAuth(deps, testLogger())
	if err != nil {
		return nil, err
	}

	return srv.server, nil
}

func deps() MCPDeps {
	f := &fakeDeps{}

	return MCPDeps{
		Remember:   f,
		Recall:     f,
		GetFact:    f,
		RemoveFact: f,
	}
}

func TestToolDescriptions(t *testing.T) {
	cs, ctx := connect(t, withAuth, deps())

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, tool := range res.Tools {
		got[tool.Name] = tool.Description
		if words := len(strings.Fields(tool.Description)); words == 0 || words > 20 {
			t.Errorf("%s description has %d words, want 1..20", tool.Name, words)
		}
	}

	for _, want := range []string{"remember", "recall", "read_fact", "remove_fact_by_id"} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing tool %q, got %v", want, keys(got))
		}
	}

	if _, ok := got["get_fact_by_id"]; ok {
		t.Error("stale tool name get_fact_by_id is still exposed")
	}

	// recall and read_fact must reference each other, otherwise the model
	// cannot tell "short summary" and "full text" apart.
	if !strings.Contains(got["recall"], "read_fact") {
		t.Error("recall description does not reference read_fact")
	}

	if !strings.Contains(got["recall"], "top-K") {
		t.Error("recall description does not state it returns top-K only")
	}

	if strings.Contains(got["recall"], "Recall all") {
		t.Error("recall description still claims to return all facts")
	}

	if !strings.Contains(got["read_fact"], "recall") {
		t.Error("read_fact description does not reference recall")
	}

	if !strings.Contains(got["read_fact"], "full") {
		t.Error("read_fact description does not state it returns the full text")
	}
}

func TestInstructionsMentionEveryTool(t *testing.T) {
	cs, _ := connect(t, withAuth, deps())

	init := cs.InitializeResult()
	if init == nil {
		t.Fatal("no initialize result")
	}

	instr := init.Instructions

	for _, want := range []string{"remember", "recall", "read_fact", "remove_fact_by_id"} {
		if !strings.Contains(instr, want) {
			t.Errorf("instructions do not mention %q", want)
		}
	}

	if strings.Contains(instr, "getFactByID") {
		t.Error("instructions still use the old camelCase tool name getFactByID")
	}

	// The guidance that prevented the "recall returns everything" mistake.
	for _, want := range []string{"top-K", "does not mean memory is empty", "read_fact"} {
		if !strings.Contains(instr, want) {
			t.Errorf("instructions lost key guidance: %q", want)
		}
	}
}

func TestInstructionsCoverProactiveRecall(t *testing.T) {
	cs, _ := connect(t, withAuth, deps())

	init := cs.InitializeResult()
	if init == nil {
		t.Fatal("no initialize result")
	}

	instr := init.Instructions

	// Proactive recall is wanted, so the trigger must be stated explicitly
	// rather than left to the user to ask for.
	if !strings.Contains(instr, "When to recall without being asked") {
		t.Error("instructions do not explain when to recall proactively")
	}

	// The anti-triggers matter as much as the triggers: recall on every
	// message pollutes the answer and costs a round trip.
	for _, want := range []string{"Don't recall when", "general question about programming", "every message"} {
		if !strings.Contains(instr, want) {
			t.Errorf("instructions lack a recall anti-trigger: %q", want)
		}
	}

	if !strings.Contains(instr, "weakest source of truth") {
		t.Error("instructions rank memory below the current context")
	}
}

func TestInstructionsCoverConflictResolution(t *testing.T) {
	cs, _ := connect(t, withAuth, deps())

	init := cs.InitializeResult()
	if init == nil {
		t.Fatal("no initialize result")
	}

	instr := init.Instructions
	for _, want := range []string{
		"Save the new fact with `remember` first",
		"After a successful save, use `recall`",
		"Use `read_fact` to check their full contents",
		"full or partial conflict",
		"proactively show the user the old fact's ID and relevant contents",
		"still-valid information",
		"ask whether to delete the old fact",
		"only after the user explicitly requests its deletion",
		"confirming deletion of the identified fact",
		"not permission to delete",
		"Declining or not answering leaves both facts stored",
		"without asking again",
		"These limits on contextual recall do not apply to the conflict check",
	} {
		if !strings.Contains(instr, want) {
			t.Errorf("instructions lost conflict-resolution guidance: %q", want)
		}
	}

	if strings.Contains(instr, "or agrees that a stored fact is wrong") {
		t.Error("instructions still allow deletion merely because the user says a fact is wrong")
	}
}

func TestToolDescriptionsPreserveEssentialGuidance(t *testing.T) {
	cs, ctx := connect(t, withAuth, deps())

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	descriptions := map[string]string{}
	for _, tool := range res.Tools {
		descriptions[tool.Name] = tool.Description
	}

	guidance := map[string][]string{
		"remember": {
			"user-provided fact",
			"short summary",
			"recall and read_fact",
			"check conflicts",
		},
		"recall": {
			"Semantic search",
			"not inventory",
			"summaries, never full text",
		},
		"read_fact": {
			"full text and summary",
			"conflicts before proposing deletion",
		},
		"remove_fact_by_id": {
			"Permanently delete",
			"only on explicit user request",
			"Ask whether to remove conflicting facts",
			"success even if absent",
		},
	}
	for name, fragments := range guidance {
		t.Run(name, func(t *testing.T) {
			description, ok := descriptions[name]
			if !ok {
				t.Fatalf("missing tool %q", name)
			}
			for _, want := range fragments {
				if !strings.Contains(description, want) {
					t.Errorf("description lost essential guidance: %q", want)
				}
			}
		})
	}
}

// A call made over a transport that carries no HTTP headers has no credentials.
// The two constructors are the only thing that decides what that means: with auth
// there is no user to act for, without auth the call goes through as it always did.
func TestCallWithoutHeaders(t *testing.T) {
	tests := []struct {
		name        string
		build       buildFunc
		wantRefused bool
	}{
		{name: "with auth", build: withAuth, wantRefused: true},
		{name: "without auth", build: withoutAuth, wantRefused: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, ctx := connect(t, tt.build, deps())

			res, err := cs.CallTool(ctx, &sdk.CallToolParams{
				Name:      "recall",
				Arguments: map[string]any{"query": "anything"},
			})
			if err != nil {
				t.Fatalf("CallTool: %v", err)
			}

			if res.IsError != tt.wantRefused {
				t.Errorf("IsError = %v, want %v", res.IsError, tt.wantRefused)
			}
		})
	}
}

// Both constructors must expose the same tools: the mode changes who may call
// them, never what exists.
func TestBothConstructorsExposeEveryTool(t *testing.T) {
	builds := map[string]buildFunc{"with auth": withAuth, "without auth": withoutAuth}

	for name, build := range builds {
		t.Run(name, func(t *testing.T) {
			cs, ctx := connect(t, build, deps())

			res, err := cs.ListTools(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}

			names := make([]string, 0, len(res.Tools))
			for _, tool := range res.Tools {
				names = append(names, tool.Name)
			}

			if len(names) != 4 {
				t.Errorf("tools = %v, want 4", names)
			}
		})
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
