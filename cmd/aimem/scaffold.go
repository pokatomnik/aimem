package main

import (
	"context"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp"

	config "github.com/pokatomnik/gomem/cmd/aimem/config"
	aInputPackage "github.com/pokatomnik/gomem/internal/adapters/input"
	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	aMCPHTTPPackage "github.com/pokatomnik/gomem/internal/adapters/servers/http"
	aMCPStdIOPackage "github.com/pokatomnik/gomem/internal/adapters/servers/stdio"
	cmdConsumePackage "github.com/pokatomnik/gomem/internal/controllers/consume"
	cmdDumpPackage "github.com/pokatomnik/gomem/internal/controllers/dump"
	cmdGetFactPackage "github.com/pokatomnik/gomem/internal/controllers/getfact"
	cmdMCPHTTPPackage "github.com/pokatomnik/gomem/internal/controllers/http"
	cmdMCPPackage "github.com/pokatomnik/gomem/internal/controllers/mcp"
	cmdRecallPackage "github.com/pokatomnik/gomem/internal/controllers/recall"
	cmdRememberPackage "github.com/pokatomnik/gomem/internal/controllers/remember"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/root"
	cmdMCPStdIOPackage "github.com/pokatomnik/gomem/internal/controllers/stdio"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	ucConsumePackage "github.com/pokatomnik/gomem/internal/use_cases/consume"
	ucDumpPackage "github.com/pokatomnik/gomem/internal/use_cases/dump"
	ucGetFactPackage "github.com/pokatomnik/gomem/internal/use_cases/getfact"
	ucMCPHTTPPackage "github.com/pokatomnik/gomem/internal/use_cases/mcphttp"
	ucMCPStdIOPackage "github.com/pokatomnik/gomem/internal/use_cases/mcpstdio"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
)

func scaffold(ctx context.Context) (*cobra.Command, *rKVPackage.KV, error) {
	cfg, err := config.InitConfig()
	if err != nil {
		return nil, nil, err
	}

	client := http.DefaultClient

	db, err := rKVPackage.Open(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	aOllama := aOllamaPackage.New(client, cfg)
	aInput := aInputPackage.New(os.Stdin)

	rec := ucRecallPackage.New(aOllama, db, cfg)
	fac := ucGetFactPackage.New(db)
	rem := ucRememberPackage.New(aOllama, db)

	mcpDeps := mcp.MCPDeps{
		Remember: rem,
		Recall:   rec,
		GetFact:  fac,
	}

	aMCPHTTP := aMCPHTTPPackage.New(mcpDeps)
	aMCPStdIO := aMCPStdIOPackage.New(mcpDeps)

	con := ucConsumePackage.New(aOllama, db, aInput)
	dum := ucDumpPackage.New(db)
	mcphttp := ucMCPHTTPPackage.New(aMCPHTTP)
	mcpstdio := ucMCPStdIOPackage.New(aMCPStdIO)

	remember := cmdRememberPackage.RememberCmd(rem)
	recall := cmdRecallPackage.RecallCmd(rec)
	getFact := cmdGetFactPackage.GetFactCmd(fac)
	dump := cmdDumpPackage.DumpCmd(dum)
	consume := cmdConsumePackage.Consume(con)
	mcpStdIO := cmdMCPStdIOPackage.NewMCPStdIOController(mcpstdio)
	mcpHTTP := cmdMCPHTTPPackage.NewMCPHTTPController(cfg, mcphttp)
	mcp := cmdMCPPackage.NewMCPController(mcpStdIO, mcpHTTP)
	root := cmdRootPackage.RootCmd(remember, recall, getFact, dump, consume, mcp)

	return root, db, nil
}
