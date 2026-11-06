package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/pokatomnik/gomem/internal/shared/logging"
)

func openLogFile(path string) (*os.File, func(), error) {
	if path == "" {
		return os.Stderr, func() {}, nil
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file %q: %w", path, err)
	}

	return file, func() { _ = file.Close() }, nil
}

// newLogger builds the process logger over the output file selected by main.
func newLogger(w io.Writer, level string) *slog.Logger {
	return logging.New(w, logging.ParseLevel(level))
}
