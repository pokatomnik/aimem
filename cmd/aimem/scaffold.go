package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/pokatomnik/gomem/internal/adapters/servers/mcp"
	"github.com/pokatomnik/gomem/internal/shared/colors"

	config "github.com/pokatomnik/gomem/cmd/aimem/config"
	aInputPackage "github.com/pokatomnik/gomem/internal/adapters/input"
	aOllamaPackage "github.com/pokatomnik/gomem/internal/adapters/ollama"
	aMCPHTTPPackage "github.com/pokatomnik/gomem/internal/adapters/servers/http"
	cmdInitPackage "github.com/pokatomnik/gomem/internal/controllers/cli/init"
	cmdRootPackage "github.com/pokatomnik/gomem/internal/controllers/cli/root"
	cmdWebCLIPackage "github.com/pokatomnik/gomem/internal/controllers/cli/web"
	cmdWebPackage "github.com/pokatomnik/gomem/internal/controllers/web"
	rKVPackage "github.com/pokatomnik/gomem/internal/repositories/kv"
	rQdrantPackage "github.com/pokatomnik/gomem/internal/repositories/qdrant"
	rUsersPackage "github.com/pokatomnik/gomem/internal/repositories/users"
	ucGetFactPackage "github.com/pokatomnik/gomem/internal/use_cases/getfact"
	ucInitPackage "github.com/pokatomnik/gomem/internal/use_cases/init"
	ucRecallPackage "github.com/pokatomnik/gomem/internal/use_cases/recall"
	ucRememberPackage "github.com/pokatomnik/gomem/internal/use_cases/remember"
	ucRemoveFactPackage "github.com/pokatomnik/gomem/internal/use_cases/removefact"
)

func scaffold(ctx context.Context) (*cobra.Command, error) {
	configPath, err := config.ConfigPath()
	if err != nil {
		return nil, err
	}

	aInput := aInputPackage.New(os.Stdin)
	init := ucInitPackage.New(configPath, config.DefaultConfig().String(), aInput)

	cfg, err := config.GetConfig()
	if err != nil {
		initialized, dumpErr := init.Dump()
		if dumpErr != nil {
			log.Fatalf("Failed to initialize configuration: %s\n", dumpErr.Error())
		}

		if initialized {
			log.Fatal(colors.Primary().Sprintf("Initial configuration dumped. Please set It up: %s\n", configPath))
		}

		// The file exists but could not be read or parsed, and the user declined to
		// overwrite it. Every constructor below dereferences cfg, so carrying on
		// would panic on a nil pointer; report the original failure instead.
		log.Fatalf("Failed to read configuration: %s\n", err.Error())
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
	if cfg.IsQdrantConfigured() {
		rec = ucRecallPackage.New(aOllama, qdrantDB, cfg)
		fac = ucGetFactPackage.New(qdrantDB)
		rem = ucRememberPackage.New(aOllama, qdrantDB)
		del = ucRemoveFactPackage.New(qdrantDB)
	}

	mcpDeps := mcp.MCPDeps{
		Remember:   rem,
		Recall:     rec,
		GetFact:    fac,
		RemoveFact: del,
	}

	// One branch, both servers. The constructors are chosen from the same answer, so
	// the web API and the MCP tools cannot end up in different modes.
	var (
		aMCPHTTP      *aMCPHTTPPackage.HTTP
		webController *cmdWebPackage.Controller
	)

	if authRequired {
		aMCPHTTP, err = aMCPHTTPPackage.TryCreateWithAuth(mcpDeps, users)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithAuth(cfg, rem, rec, fac, del, aMCPHTTP, users)
	} else {
		aMCPHTTP, err = aMCPHTTPPackage.TryCreateWithoutAuth(mcpDeps)
		if err != nil {
			return nil, err
		}

		webController = cmdWebPackage.NewWithoutAuth(cfg, rem, rec, fac, del, aMCPHTTP, users)
	}

	web := cmdWebCLIPackage.NewWebController(webController)

	cmdInitPackage := cmdInitPackage.GetInitCmd(init)

	root := cmdRootPackage.RootCmd(web, cmdInitPackage)

	return root, nil
}
