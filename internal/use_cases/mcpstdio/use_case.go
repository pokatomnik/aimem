package mcpstdio

import "context"

type (
	Runner interface {
		Run(ctx context.Context) error
	}
	StdIO struct {
		runner Runner
	}
)

func New(runner Runner) *StdIO {
	return &StdIO{runner}
}

func (s *StdIO) Run(ctx context.Context) error {
	return s.runner.Run(ctx)
}
