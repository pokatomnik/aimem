package http

import (
	"cmp"
	"context"

	"github.com/spf13/cobra"
)

const defaultHTTPAddress = "127.0.0.1:8080"

type (
	Config interface {
		GetMCPHTTPHost() string
	}
	serverRunner interface {
		Run(ctx context.Context, host string) error
	}
)

func NewMCPHTTPController(config Config, serverRunner serverRunner, cmds ...*cobra.Command) *cobra.Command {
	runMCPHTTP := &cobra.Command{
		Use:           "http",
		Short:         "Run MCP on HTTP",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       "cql mcp http",
		RunE: func(cmd *cobra.Command, args []string) error {
			host := cmp.Or(config.GetMCPHTTPHost(), defaultHTTPAddress)
			return serverRunner.Run(cmd.Context(), host)
		},
	}

	runMCPHTTP.AddCommand(cmds...)

	return runMCPHTTP
}
