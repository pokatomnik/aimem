package removefact

import (
	"context"

	"github.com/pokatomnik/gomem/internal/shared/colors"
	"github.com/spf13/cobra"
)

type removeFact interface {
	RemoveFactByID(ctx context.Context, id string) error
}

func RemoveFactCmd(r removeFact, nested ...*cobra.Command) *cobra.Command {
	removeFactCmd := &cobra.Command{
		Use:           "remove <factID>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Remove fact by ID",
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := r.RemoveFactByID(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			colors.Primary().Println("Removed")

			return nil
		},
	}

	removeFactCmd.AddCommand(nested...)

	return removeFactCmd
}
