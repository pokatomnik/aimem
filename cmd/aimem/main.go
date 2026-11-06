package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(run())
}

// run keeps the exit code in one place, so every deferred cleanup — closing the
// log file and stopping signal handling — runs before the process ends. Calling
// log.Fatal or os.Exit from the body would skip those defers, and the final error
// would bypass the configured logger: log.Fatal always writes to stderr and
// ignores LOG_FILE, LOG_LEVEL and the source/context attributes.
func run() int {
	// The signal context comes first so that startup and serving share one
	// cancellation: scaffold takes it while resolving the authentication mode, and
	// ExecuteContext hands the same context to whichever command runs.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
	)
	defer stop()

	cfg, err := GetConfig()
	if err != nil {
		// No logger yet, and no configured destination to put this in.
		fmt.Fprintf(os.Stderr, "Configuration init failed: %v\n", err)

		return 1
	}

	logger, closeLog, err := newLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logging init failed: %v\n", err)

		return 1
	}
	defer closeLog()

	logger.Info("starting aimem")

	root, err := scaffold(ctx, cfg, logger)
	if err != nil {
		logger.Error("startup failed", "error", err)

		return 1
	}

	err = root.ExecuteContext(ctx)
	if errors.Is(err, context.Canceled) {
		logger.Info("cancelled")

		return 0
	}

	if err != nil {
		logger.Error("command failed", "error", err)

		return 1
	}

	logger.Info("stopped")

	return 0
}
