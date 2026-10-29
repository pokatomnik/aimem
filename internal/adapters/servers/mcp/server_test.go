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
// it only has to be non-nil.
type fakeDeps struct{}

func (f *fakeDeps) Remember(ctx context.Context, userID, s string) (string, error) {
	return "summary of " + s, nil
}

func (f *fakeDeps) Recall(ctx context.Context, userID, q string) ([]fact.FactListView, error) {
	return []fact.FactListView{
		{ID: "id-1", CreatedAt: "2026-01-01", Summary: "short summary one"},
	}, nil
}

func (f *fakeDeps) GetFact(ctx context.Context, userID, id string) (fact.FactFullView, error) {
	return fact.FactFullView{ID: id}, nil
}

func (f *fakeDeps) RemoveFactByID(ctx context.Context, userID, id string) error {
	return nil
}

func (f *fakeDeps) GetUserByToken(ctx context.Context, token string) (user.User, error) {
	return user.User{ID: "u-" + token}, nil
}

func connect(t *testing.T, deps MCPDeps) (*sdk.ClientSession, context.Context) {
	t.Helper()

	server, err := CreateServer(deps)
	if err != nil {
		t.Fatal(err)
	}

	serverT, clientT := sdk.NewInMemoryTransports()
	ctx := context.Background()

	go func() { _ = server.Run(ctx, serverT) }()

	client := sdk.NewClient(&sdk.Implementation{Name: "test"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { cs.Close() })

	return cs, ctx
}

func deps() MCPDeps {
	f := &fakeDeps{}

	return MCPDeps{
		Remember:       f,
		Recall:         f,
		GetFact:        f,
		RemoveFact:     f,
		Authentication: f,
	}
}

func TestToolDescriptions(t *testing.T) {
	cs, ctx := connect(t, deps())

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, tool := range res.Tools {
		got[tool.Name] = tool.Description
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
	cs, _ := connect(t, deps())

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
	cs, _ := connect(t, deps())

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

// A call made over a transport that carries no HTTP headers must be refused, not
// answered: there are no credentials, so there is no user to read or write for.
func TestCallWithoutHeadersIsRefused(t *testing.T) {
	cs, ctx := connect(t, deps())

	res, err := cs.CallTool(ctx, &sdk.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": "anything"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("call succeeded without credentials")
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
