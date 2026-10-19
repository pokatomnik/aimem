package main

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"

	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	cmdRecallPackage "github.com/pokatomnik/gomem/internal/controllers/recall"
	cmdRememberPackage "github.com/pokatomnik/gomem/internal/controllers/remember"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/root"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
)

func scaffold(ctx context.Context) (*cobra.Command, *rKVPackage.KV, error) {
	config, err := InitConfig()
	if err != nil {
		return nil, nil, err
	}

	client := http.DefaultClient

	ollama := aOllamaPackage.New(client, config)

	db, err := rKVPackage.Open(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	rem := ucRememberPackage.New(ollama, db)
	rec := ucRecallPackage.New(ollama, db, config)

	remember := cmdRememberPackage.RememberCmd(rem)
	recall := cmdRecallPackage.RecallCmd(rec)
	root := cmdRootPackage.RootCmd(remember, recall)

	return root, db, nil
}
