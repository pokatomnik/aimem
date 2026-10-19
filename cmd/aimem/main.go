package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	root, db, err := scaffold(ctx)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
