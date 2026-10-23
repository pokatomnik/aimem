package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pokatomnik/gomem/internal/entities/fact"
)

type fakeDeps struct {
	recalled string
	readID   string
}

func (f *fakeDeps) Remember(ctx context.Context, s string) (string, error) {
	return "summary of " + s, nil
}

func (f *fakeDeps) Recall(ctx context.Context, q string) ([]fact.FactListView, error) {
	f.recalled = q
	return []fact.FactListView{
		{ID: "id-1", CreatedAt: "2026-01-01", Summary: "short summary one"},
		{ID: "id-2", CreatedAt: "2026-01-02", Summary: "short summary two"},
	}, nil
}

func (f *fakeDeps) GetFact(ctx context.Context, id string) (fact.FactFullView, error) {
	f.readID = id
	return fact.FactFullView{
		ID:        id,
		CreatedAt: "2026-01-01",
		Summary:   "short",
		Content:   "the full stored text",
	}, nil
}

func (f *fakeDeps) RemoveFactByID(ctx context.Context, id string) error {
	return nil
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

func deps(f *fakeDeps) MCPDeps {
	return MCPDeps{Remember: f, Recall: f, GetFact: f, RemoveFact: f}
}

func TestToolDescriptions(t *testing.T) {
	cs, ctx := connect(t, deps(&fakeDeps{}))

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
	cs, _ := connect(t, deps(&fakeDeps{}))

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

func TestRecallReturnsSummariesOnly(t *testing.T) {
	f := &fakeDeps{}
	cs, ctx := connect(t, deps(f))

	res, err := cs.CallTool(ctx, &sdk.CallToolParams{
		Name:      "recall",
		Arguments: map[string]any{"query": "anything"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.IsError {
		t.Fatalf("recall failed: %v", res.Content)
	}

	b, _ := json.Marshal(res.StructuredContent)

	if strings.Contains(string(b), "the full stored text") {
		t.Error("recall leaked the full fact text; it must return summaries only")
	}

	if !strings.Contains(string(b), "short summary one") {
		t.Errorf("recall did not return summaries: %s", b)
	}

	if f.recalled != "anything" {
		t.Errorf("query not forwarded, got %q", f.recalled)
	}
}

func TestReadFactReturnsFullText(t *testing.T) {
	f := &fakeDeps{}
	cs, ctx := connect(t, deps(f))

	res, err := cs.CallTool(ctx, &sdk.CallToolParams{
		Name:      "read_fact",
		Arguments: map[string]any{"id": "id-1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.IsError {
		t.Fatalf("read_fact failed: %v", res.Content)
	}

	b, _ := json.Marshal(res.StructuredContent)

	if !strings.Contains(string(b), "the full stored text") {
		t.Errorf("read_fact did not return the full text: %s", b)
	}

	if f.readID != "id-1" {
		t.Errorf("id not forwarded, got %q", f.readID)
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
