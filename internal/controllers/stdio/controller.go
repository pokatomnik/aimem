package stdio

import (
	"context"

	"github.com/spf13/cobra"
)

type serverRunner interface {
	Run(ctx context.Context) error
}

func NewMCPStdIOController(serverRunner serverRunner, cmds ...*cobra.Command) *cobra.Command {
	runMCPStdIO := &cobra.Command{
		Use:           "stdio",
		Short:         "Run MCP on stdio",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serverRunner.Run(cmd.Context())
		},
	}

	runMCPStdIO.AddCommand(cmds...)

	return runMCPStdIO
}
