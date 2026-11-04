package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"
)

func main() {
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

	root, err := scaffold(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = root.ExecuteContext(ctx)
	if errors.Is(err, context.Canceled) {
		fmt.Println("Cancelled")

		return
	}

	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}
