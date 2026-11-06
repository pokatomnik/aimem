package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := map[string]slog.Level{
		LevelDebug:   slog.LevelDebug,
		"DEBUG":      slog.LevelDebug,
		LevelInfo:    slog.LevelInfo,
		"":           slog.LevelInfo,
		LevelWarn:    slog.LevelWarn,
		"warning":    slog.LevelWarn,
		LevelError:   slog.LevelError,
		"not a name": slog.LevelInfo,
	}

	for name, want := range tests {
		if got := ParseLevel(name); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", name, got, want)
		}
	}
}

// The writer is the whole destination decision, so a logger built over a buffer
// must put every enabled record there and drop the ones below the level.
func TestNewWritesToTheWriterAndFiltersByLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := New(&buf, slog.LevelWarn)

	logger.Info("hidden")
	logger.Warn("shown", "key", "value")

	out := buf.String()

	if strings.Contains(out, "hidden") {
		t.Errorf("output contains a below-level record: %q", out)
	}

	if !strings.Contains(out, "shown") {
		t.Errorf("output does not contain the warn record: %q", out)
	}

	if !strings.Contains(out, "key=value") {
		t.Errorf("output lost the attribute: %q", out)
	}

	if !strings.Contains(out, "source=") {
		t.Errorf("output does not contain the source location: %q", out)
	}
}

func TestDiscardWritesNowhere(t *testing.T) {
	t.Parallel()

	// A nil writer would discard anyway; Discard is the named form of that, so
	// the test pins that it can be called at any level without a panic.
	Discard().Error("ignored", "error", "boom")
}
