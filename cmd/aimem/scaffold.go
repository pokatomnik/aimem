package main

import (
	"context"
	"net/http"

	"github.com/pokatomnik/gomem/internal/adapters/ollama"
	"github.com/pokatomnik/gomem/internal/repositories/kv"
	"github.com/pokatomnik/gomem/internal/use_cases/recall"
	"github.com/pokatomnik/gomem/internal/use_cases/remember"
)

func scaffold(ctx context.Context) (*remember.Remember, *recall.Recall, error) {
	config, err := InitConfig()
	if err != nil {
		return nil, nil, err
	}

	client := http.DefaultClient

	ollama := ollama.New(client, config)

	db, err := kv.Open(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	rem := remember.New(ollama, db)
	rec := recall.New(ollama, db, config)

	return rem, rec, nil
}
