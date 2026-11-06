package qdrant

import (
	"log/slog"

	"github.com/pokatomnik/gomem/internal/shared/logging"
)

// testLogger is the production logger's method set with a writer that discards
// everything, so the tests stay free of I/O.
func testLogger() *slog.Logger {
	return logging.Discard()
}
