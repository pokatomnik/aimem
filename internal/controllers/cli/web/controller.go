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
		onStart func(where string),
	) error
}

func NewWebController(serverRunner serverRunner, cmds ...*cobra.Command) *cobra.Command {
	runWebController := &cobra.Command{
		Use:           "web",
		Short:         "Run web-based search engine",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serverRunner.Run(
				cmd.Context(),
				&server.ListenAndServeOptions{},
				func(where string) {
					colors.Primary().Fprintf(os.Stderr, "Server started at %s\n", where)
				},
			)
		},
	}

	runWebController.AddCommand(cmds...)

	return runWebController
}
