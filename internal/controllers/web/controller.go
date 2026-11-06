package web

import (
	"cmp"
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/pokatomnik/gomem/internal/controllers/web/server"
	httpServerExtPkg "github.com/pokatomnik/gomem/internal/shared/server"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen@v1.24.0 --target server -package server --clean openapi.yaml

type Controller struct {
	cfg            Config
	remember       Remember
	recall         Recall
	getFact        GetFact
	removeFact     RemoveFact
	mcpHandler     http.Handler
	authentication Authentication
	// authRequired is fixed by the constructor. It decides whether a request needs a
	// resolved user, and is baked into the served page so the API and the UI cannot
	// disagree.
	authRequired bool
	// logger records startup and every failure that becomes a response. It is
	// injected rather than created here, so the destination and level are decided
	// once at the composition root.
	logger logger
}

const defaultWebHost = "127.0.0.1:8888"

// NewWithAuth serves a deployment where every request must be authenticated, i.e.
// no user can be inferred without a token.
func NewWithAuth(
	cfg Config,
	remember Remember,
	recall Recall,
	getFact GetFact,
	removeFact RemoveFact,
	mcpHandler http.Handler,
	authentication Authentication,
	logger logger,
) *Controller {
	return newController(cfg, remember, recall, getFact, removeFact, mcpHandler, authentication, logger, true)
}

// NewWithoutAuth serves a deployment where the header carries no meaning: one
// local operator, and a store that ignores the user ID entirely.
func NewWithoutAuth(
	cfg Config,
	remember Remember,
	recall Recall,
	getFact GetFact,
	removeFact RemoveFact,
	mcpHandler http.Handler,
	authentication Authentication,
	logger logger,
) *Controller {
	return newController(cfg, remember, recall, getFact, removeFact, mcpHandler, authentication, logger, false)
}

// newController is the single place the two constructors share, so the fields they
// set cannot drift apart.
func newController(
	cfg Config,
	remember Remember,
	recall Recall,
	getFact GetFact,
	removeFact RemoveFact,
	mcpHandler http.Handler,
	authentication Authentication,
	logger logger,
	authRequired bool,
) *Controller {
	return &Controller{
		cfg:            cfg,
		remember:       remember,
		recall:         recall,
		getFact:        getFact,
		removeFact:     removeFact,
		mcpHandler:     mcpHandler,
		authentication: authentication,
		authRequired:   authRequired,
		logger:         logger,
	}
}

func (c *Controller) HandleBearerAuth(
	ctx context.Context,
	operationName server.OperationName,
	t server.BearerAuth,
) (context.Context, error) {
	// With authentication off the header carries no meaning: the embedded
	// backend serves one local operator and ignores userID. Accepting the request
	// without resolving anything keeps the API open, exactly as it was before
	// tokens existed.
	if !c.authRequired {
		return ctx, nil
	}

	u, err := c.authentication.GetUserByToken(ctx, t.Token)
	if err != nil {
		c.logger.Warn("bearer token rejected", "operation", string(operationName), "error", err)

		return ctx, err
	}

	return context.WithValue(ctx, userIDKey{}, u.ID), nil
}

func (c *Controller) Run(
	ctx context.Context,
	options *httpServerExtPkg.ListenAndServeOptions,
	onStartListening func(where string),
) error {
	host := cmp.Or(c.cfg.GetWebHTTPHost(), defaultWebHost)

	c.logger.Info("web server starting", "address", host)

	opts := ServerOptions(c.logger)
	api, err := server.NewServer(&Methods{
		remember:     c.remember,
		recall:       c.recall,
		getFact:      c.getFact,
		removeFact:   c.removeFact,
		authRequired: c.authRequired,
		logger:       c.logger,
	}, c, opts...)
	if err != nil {
		c.logger.Error("create web API server failed", "address", host, "error", err)

		return err
	}

	spa, err := newSPAHandler(c.authRequired, c.logger)
	if err != nil {
		c.logger.Error("create SPA handler failed", "auth_required", c.authRequired, "error", err)

		return err
	}

	ctxServer := httpServerExtPkg.NewHTTPServerExt(&http.Server{
		Addr:    host,
		Handler: newHandler(api, spa, c.mcpHandler),
		BaseContext: func(l net.Listener) context.Context {
			onStartListening(l.Addr().String())

			return ctx
		},
	})

	err = ctxServer.ListenAndServeContext(ctx, options)
	if errors.Is(err, http.ErrServerClosed) {
		c.logger.Info("web server stopped", "address", host)

		return nil
	}

	if err != nil {
		c.logger.Error("web server failed", "address", host, "error", err)

		return err
	}

	return nil
}
