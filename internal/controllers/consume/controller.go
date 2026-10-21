package consume

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

type consumer interface {
	Consume(ctx context.Context) []error
}

func Consume(c consumer, nested ...*cobra.Command) *cobra.Command {
	consumeCmd := &cobra.Command{
		Use:           "import",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Imports facts from stdin",
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			errs := c.Consume(cmd.Context())

			if len(errs) != 0 {
				return errors.Join(errs...)
			}

			return nil
		},
	}

	consumeCmd.AddCommand(nested...)

	return consumeCmd
}
