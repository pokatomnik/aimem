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
}

const defaultWebHost = "127.0.0.1:8888"

func New(
	cfg Config,
	remember Remember,
	recall Recall,
	getFact GetFact,
	removeFact RemoveFact,
	mcpHandler http.Handler,
	authentication Authentication,
) *Controller {
	return &Controller{
		cfg:            cfg,
		remember:       remember,
		recall:         recall,
		getFact:        getFact,
		removeFact:     removeFact,
		mcpHandler:     mcpHandler,
		authentication: authentication,
	}
}

func (c *Controller) HandleBearerAuth(
	ctx context.Context,
	operationName server.OperationName,
	t server.BearerAuth,
) (context.Context, error) {
	u, err := c.authentication.GetUserByToken(ctx, t.Token)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, userIDKey{}, u.ID), nil
}

func (c *Controller) Run(
	ctx context.Context,
	options *httpServerExtPkg.ListenAndServeOptions,
) error {
	host := cmp.Or(c.cfg.GetWebHTTPHost(), defaultWebHost)

	opts := ServerOptions()
	api, err := server.NewServer(&Methods{
		remember:   c.remember,
		recall:     c.recall,
		getFact:    c.getFact,
		removeFact: c.removeFact,
	}, c, opts...)
	if err != nil {
		return err
	}

	spa, err := newSPAHandler()
	if err != nil {
		return err
	}

	ctxServer := httpServerExtPkg.NewHTTPServerExt(&http.Server{
		Addr:    host,
		Handler: newHandler(api, spa, c.mcpHandler),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	})

	err = ctxServer.ListenAndServeContext(ctx, options)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}
