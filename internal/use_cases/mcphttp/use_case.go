package mcphttp

import (
	"context"

	"github.com/pokatomnik/gomem/internal/shared/server"
)

type (
	Runner interface {
		Run(ctx context.Context, host string, options *server.ListenAndServeOptions) error
	}
	HTTP struct {
		runner Runner
	}
)

func New(runner Runner) *HTTP {
	return &HTTP{runner: runner}
}

func (h *HTTP) Run(ctx context.Context, host string) error {
	return h.runner.Run(ctx, host, &server.ListenAndServeOptions{})
}
