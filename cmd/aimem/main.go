package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()
	// remember, _, err := scaffold(ctx)
	_, recall, err := scaffold(ctx)
	if err != nil {
		log.Fatal(err)
	}

	facts, err := recall.Recall(ctx, "Алкоголь")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(facts)

	// err = remember.Remember(ctx, "Детям нельзя пить водку")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Println("Done")
}
