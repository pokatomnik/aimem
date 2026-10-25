package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pokatomnik/gomem/cmd/aimem/config"
)

func main() {
	ctx := context.Background()

	root, err := scaffold()
	if err != nil {
		fmt.Fprintln(os.Stderr, config.Example())
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
	)
	defer stop()

	err = root.ExecuteContext(ctx)
	if errors.Is(err, context.Canceled) {
		fmt.Println("Cancelled")

		return
	}

	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}
