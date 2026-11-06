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
// os.Exit from the body would skip those defers.
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

	logFile, closeLogFile, err := openLogFile(cfg.GetLogFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logging init failed: %v\n", err)

		return 1
	}
	defer closeLogFile()

	root, err := scaffold(ctx, cfg, logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)

		return 1
	}

	err = root.ExecuteContext(ctx)
	if errors.Is(err, context.Canceled) {
		return 0
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)

		return 1
	}

	return 0
}
