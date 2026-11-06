package main

import (
	"context"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"

	aEmbedderPackage "github.com/pokatomnik/gomem/internal/adapters/embeder"
	aLLMPackage "github.com/pokatomnik/gomem/internal/adapters/llm"
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

func scaffold(ctx context.Context, cfg *Config, logFile *os.File) (*cobra.Command, error) {
	logger := newLogger(logFile, cfg.GetLogLevel())
	client := http.DefaultClient

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	embeddedDB := rKVPackage.New(cfg, path.Join(userHome, ".aimem"), logger)
	qdrantDB := rQdrantPackage.New(cfg, client, logger)
	users := rUsersPackage.New(cfg, logger)

	// One decision, taken once and expressed in the constructors below: whether this
	// deployment is per-user at all. It follows the storage backend, not the
	// configured users — Qdrant partitions by user and therefore needs a token, the
	// embedded index serves one local operator and ignores the user ID. With Qdrant
	// and no user in AIMEM_USERS_JSON this call fails, which stops the app before it
	// serves anything a client could not get past.
	authRequired, err := users.RequireAuthentication(ctx)
	if err != nil {
		return nil, err
	}

	embedder := aEmbedderPackage.New(logger)
	llm := aLLMPackage.New(cfg, logger)

	rec := ucRecallPackage.New(embedder, embeddedDB, cfg, logger)
	fac := ucGetFactPackage.New(embeddedDB, logger)
	rem := ucRememberPackage.New(embedder, llm, embeddedDB, logger)
	del := ucRemoveFactPackage.New(embeddedDB, logger)
	if cfg.RequireAuth() {
		rec = ucRecallPackage.New(embedder, qdrantDB, cfg, logger)
		fac = ucGetFactPackage.New(qdrantDB, logger)
		rem = ucRememberPackage.New(embedder, llm, qdrantDB, logger)
		del = ucRemoveFactPackage.New(qdrantDB, logger)
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
		aMCP, err = aMCPPackage.NewWithAuth(mcpDeps, users, logger)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithAuth(cfg, rem, rec, fac, del, aMCP, users, logger)
	} else {
		aMCP, err = aMCPPackage.NewWithoutAuth(mcpDeps, logger)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithoutAuth(cfg, rem, rec, fac, del, aMCP, users, logger)
	}

	web := cmdWebCLIPackage.NewWebController(webController)

	root := cmdRootPackage.RootCmd(web)

	return root, nil
}
