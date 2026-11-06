package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/pokatomnik/gomem/internal/shared/logging"
)

// newLogger builds the process logger from the configuration.
//
// With no LOG_FILE the log goes to stderr, which is what a container or a
// terminal reads. When a path is set, the file is created if missing and opened
// for append, so a restart does not erase the record of the run that failed. The
// returned closer is a no-op for stderr; the caller always defers it.
func newLogger(cfg *Config) (*slog.Logger, func(), error) {
	level := logging.ParseLevel(cfg.GetLogLevel())

	path := cfg.GetLogFile()
	if path == "" {
		return logging.New(os.Stderr, level), func() {}, nil
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file %q: %w", path, err)
	}

	return logging.New(file, level), func() { _ = file.Close() }, nil
}
