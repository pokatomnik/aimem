package main

import (
	"context"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	config "github.com/pokatomnik/gomem/cmd/aimem/config"
	aInputPackage "github.com/pokatomnik/gomem/internal/adapters/input"
	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	cmdConsumePackage "github.com/pokatomnik/gomem/internal/controllers/consume"
	cmdDumpPackage "github.com/pokatomnik/gomem/internal/controllers/dump"
	cmdGetFactPackage "github.com/pokatomnik/gomem/internal/controllers/getfact"
	cmdRecallPackage "github.com/pokatomnik/gomem/internal/controllers/recall"
	cmdRememberPackage "github.com/pokatomnik/gomem/internal/controllers/remember"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/root"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	ucConsumePackage "github.com/pokatomnik/gomem/internal/use_cases/consume"
	ucDumpPackage "github.com/pokatomnik/gomem/internal/use_cases/dump"
	ucGetFactPackage "github.com/pokatomnik/gomem/internal/use_cases/getfact"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
)

func scaffold(ctx context.Context) (*cobra.Command, *rKVPackage.KV, error) {
	cfg, err := config.InitConfig()
	if err != nil {
		return nil, nil, err
	}

	client := http.DefaultClient

	ollama := aOllamaPackage.New(client, cfg)
	input := aInputPackage.New(os.Stdin)

	db, err := rKVPackage.Open(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	rem := ucRememberPackage.New(ollama, db)
	rec := ucRecallPackage.New(ollama, db, cfg)
	fac := ucGetFactPackage.New(db)
	dum := ucDumpPackage.New(db)
	con := ucConsumePackage.New(ollama, db, input)

	remember := cmdRememberPackage.RememberCmd(rem)
	recall := cmdRecallPackage.RecallCmd(rec)
	getFact := cmdGetFactPackage.GetFactCmd(fac)
	dump := cmdDumpPackage.DumpCmd(dum)
	cmdConsume := cmdConsumePackage.Consume(con)
	root := cmdRootPackage.RootCmd(remember, recall, getFact, dump, cmdConsume)

	return root, db, nil
}
