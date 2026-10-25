package main

import (
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp"

	config "github.com/pokatomnik/gomem/cmd/aimem/config"
	aInputPackage "github.com/pokatomnik/gomem/internal/adapters/input"
	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	aMCPHTTPPackage "github.com/pokatomnik/gomem/internal/adapters/servers/http"
	aMCPStdIOPackage "github.com/pokatomnik/gomem/internal/adapters/servers/stdio"
	cmdConsumePackage "github.com/pokatomnik/gomem/internal/controllers/cli/consume"
	cmdGetFactPackage "github.com/pokatomnik/gomem/internal/controllers/cli/getfact"
	cmdMCPHTTPPackage "github.com/pokatomnik/gomem/internal/controllers/cli/http"
	cmdMCPPackage "github.com/pokatomnik/gomem/internal/controllers/cli/mcp"
	cmdRecallPackage "github.com/pokatomnik/gomem/internal/controllers/cli/recall"
	cmdRememberPackage "github.com/pokatomnik/gomem/internal/controllers/cli/remember"
	cmdRemoveFactPackage "github.com/pokatomnik/gomem/internal/controllers/cli/removefact"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/cli/root"
	cmdMCPStdIOPackage "github.com/pokatomnik/gomem/internal/controllers/cli/stdio"
	cmdWebCLIPackage "github.com/pokatomnik/gomem/internal/controllers/cli/web"
	cmdWebPackage "github.com/pokatomnik/gomem/internal/controllers/web"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	ucConsumePackage "github.com/pokatomnik/gomem/internal/use_cases/consume"
	ucGetFactPackage "github.com/pokatomnik/gomem/internal/use_cases/getfact"
	ucMCPHTTPPackage "github.com/pokatomnik/gomem/internal/use_cases/mcphttp"
	ucMCPStdIOPackage "github.com/pokatomnik/gomem/internal/use_cases/mcpstdio"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
	ucRemoveFactPackage "github.com/pokatomnik/gomem/internal/use_cases/removefact"
)

func scaffold() (*cobra.Command, error) {
	cfg, err := config.InitConfig()
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	db := rKVPackage.New(cfg, path.Join(userHome, ".aimem"))

	aOllama := aOllamaPackage.New(client, cfg)
	aInput := aInputPackage.New(os.Stdin)

	rec := ucRecallPackage.New(aOllama, db, cfg)
	fac := ucGetFactPackage.New(db)
	rem := ucRememberPackage.New(aOllama, db)
	del := ucRemoveFactPackage.New(db)

	mcpDeps := mcp.MCPDeps{
		Remember:   rem,
		Recall:     rec,
		GetFact:    fac,
		RemoveFact: del,
	}

	aMCPHTTP := aMCPHTTPPackage.New(mcpDeps)
	aMCPStdIO := aMCPStdIOPackage.New(mcpDeps)

	con := ucConsumePackage.New(aOllama, db, aInput)
	mcphttp := ucMCPHTTPPackage.New(aMCPHTTP)
	mcpstdio := ucMCPStdIOPackage.New(aMCPStdIO)

	remember := cmdRememberPackage.RememberCmd(rem)
	recall := cmdRecallPackage.RecallCmd(rec)
	getFact := cmdGetFactPackage.GetFactCmd(fac)
	removeFact := cmdRemoveFactPackage.RemoveFactCmd(del)

	webController := cmdWebPackage.New(cfg, rem, rec, fac, del)

	consume := cmdConsumePackage.Consume(con)

	mcpStdIO := cmdMCPStdIOPackage.NewMCPStdIOController(mcpstdio)
	mcpHTTP := cmdMCPHTTPPackage.NewMCPHTTPController(cfg, mcphttp)
	mcp := cmdMCPPackage.NewMCPController(mcpStdIO, mcpHTTP)

	web := cmdWebCLIPackage.NewWebController(webController)

	root := cmdRootPackage.RootCmd(remember, recall, getFact, removeFact, consume, mcp, web)

	return root, nil
}
