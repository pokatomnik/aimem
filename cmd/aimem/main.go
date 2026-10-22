package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/pokatomnik/gomem/cmd/aimem/config"
	"github.com/pokatomnik/gomem/internal/entities/apperrors"
)

func main() {
	ctx := context.Background()

	root, db, err := scaffold(ctx)
	if errors.Is(err, apperrors.ErrConfigNotInitialized) {
		fmt.Fprintln(os.Stderr, config.Example())
	}

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
