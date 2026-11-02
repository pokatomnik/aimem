package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"

	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	aMCPPackage "github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/cli/root"
	cmdWebCLIPackage "github.com/pokatomnik/gomem/internal/controllers/cli/web"
	cmdWebPackage "github.com/pokatomnik/gomem/internal/controllers/web"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	rQdrantPackage "github.com/pokatomnik/gomem/internal/repositories/qdrant"
	rUsersPackage "github.com/pokatomnik/gomem/internal/repositories/users"
	ucGetFactPackage "github.com/pokatomnik/gomem/internal/use_cases/getfact"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
	ucRemoveFactPackage "github.com/pokatomnik/gomem/internal/use_cases/removefact"
)

func scaffold(ctx context.Context) (*cobra.Command, error) {
	cfg, err := GetConfig()
	if err != nil {
		log.Fatalf("Configuration init failed: %v", err)
	}

	client := http.DefaultClient

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	embeddedDB := rKVPackage.New(cfg, path.Join(userHome, ".aimem"))
	qdrantDB := rQdrantPackage.New(cfg, client)
	users := rUsersPackage.New(cfg)

	// One decision, taken once and expressed in the constructors below: whether this
	// deployment is per-user at all. It follows the storage backend, not the file
	// contents — Qdrant partitions by user and therefore needs a token, the embedded
	// index serves one local operator and ignores the user ID. With Qdrant and no
	// [[users]] this call fails, which stops the app before it serves anything a
	// client could not get past.
	authRequired, err := users.RequireAuthentication(ctx)
	if err != nil {
		return nil, err
	}

	aOllama := aOllamaPackage.New(client, cfg)

	rec := ucRecallPackage.New(aOllama, embeddedDB, cfg)
	fac := ucGetFactPackage.New(embeddedDB)
	rem := ucRememberPackage.New(aOllama, embeddedDB)
	del := ucRemoveFactPackage.New(embeddedDB)
	if cfg.RequireAuth() {
		rec = ucRecallPackage.New(aOllama, qdrantDB, cfg)
		fac = ucGetFactPackage.New(qdrantDB)
		rem = ucRememberPackage.New(aOllama, qdrantDB)
		del = ucRemoveFactPackage.New(qdrantDB)
	}

	mcpDeps := aMCPPackage.MCPDeps{
		Remember:   rem,
		Recall:     rec,
		GetFact:    fac,
		RemoveFact: del,
	}

	// One branch, both servers. Each pair of constructors is chosen from the same
	// answer, so the web API and the MCP endpoint cannot end up in different modes.
	var (
		aMCP          *aMCPPackage.Server
		webController *cmdWebPackage.Controller
	)

	if authRequired {
		aMCP, err = aMCPPackage.NewWithAuth(mcpDeps, users)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithAuth(cfg, rem, rec, fac, del, aMCP, users)
	} else {
		aMCP, err = aMCPPackage.NewWithoutAuth(mcpDeps)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithoutAuth(cfg, rem, rec, fac, del, aMCP, users)
	}

	web := cmdWebCLIPackage.NewWebController(webController)

	root := cmdRootPackage.RootCmd(web)

	return root, nil
}
