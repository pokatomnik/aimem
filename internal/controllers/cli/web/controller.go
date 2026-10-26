package web

import (
	"context"
	"os"

	"github.com/pokatomnik/gomem/internal/shared/colors"
	"github.com/pokatomnik/gomem/internal/shared/server"
	"github.com/spf13/cobra"
)

type serverRunner interface {
	Run(
		ctx context.Context,
		options *server.ListenAndServeOptions,
	) error
}

func NewWebController(serverRunner serverRunner, cmds ...*cobra.Command) *cobra.Command {
	runWebController := &cobra.Command{
		Use:           "web",
		Short:         "Run web-based search engine",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			colors.Primary().Fprintln(os.Stderr, "Server started")
			return serverRunner.Run(
				cmd.Context(),
				&server.ListenAndServeOptions{},
			)
		},
	}

	runWebController.AddCommand(cmds...)

	return runWebController
}
