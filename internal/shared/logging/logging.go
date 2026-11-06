package logging

import (
	"io"
	"log/slog"
	"strings"
)

// Level names accepted by ParseLevel.
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Logger is the behaviour every layer needs from the process logger.
//
// It is the method set of *log/slog.Logger, declared here because logging is a
// cross-cutting concern shared by adapters, repositories, use cases and
// controllers. Depending on this interface, never on the concrete type, is what
// lets the composition root choose the destination once and lets a test pass a
// buffer-backed logger.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ParseLevel maps LOG_LEVEL to a slog level.
//
// Matching is case-insensitive and "warning" is accepted as a synonym for
// "warn". An unknown value falls back to info rather than refusing to start: a
// typo in a log setting must not take the server down.
func ParseLevel(name string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn, "warning":
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New returns a text logger writing to w at level.
//
// The writer is the only thing that decides where the log ends up, so the same
// constructor serves stderr, a file (an *os.File is an io.Writer) and a buffer
// in a test. A nil writer discards the output instead of panicking.
//
// AddSource is on because every error record must answer where it happened, not
// only that it happened: the source attribute names the file and line, while the
// caller's attributes name the operation and the identifiers it had.
func New(w io.Writer, level slog.Level) *slog.Logger {
	if w == nil {
		w = io.Discard
	}

	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level, AddSource: true}))
}

// Discard returns a logger that writes nowhere.
func Discard() *slog.Logger {
	return New(io.Discard, slog.LevelInfo)
}
