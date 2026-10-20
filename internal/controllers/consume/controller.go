package consume

import (
	"context"

	"github.com/spf13/cobra"
)

type consumer interface {
	Consume(ctx context.Context) error
}

func Consume(c consumer, nested ...*cobra.Command) *cobra.Command {
	consumeCmd := &cobra.Command{
		Use:           "import",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Imports facts from stdin",
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Consume(cmd.Context())
		},
	}

	consumeCmd.AddCommand(nested...)

	return consumeCmd
}
