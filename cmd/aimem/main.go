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
	// The signal context comes first: scaffold resolves the authentication mode,
	// which reaches the config file, so it takes a real context rather than
	// inventing one. Interrupting startup then works the same way as interrupting
	// the server.
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
