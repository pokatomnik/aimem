package remember

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

type rememberer interface {
	Remember(ctx context.Context, fact string) (string, error)
}

func RememberCmd(r rememberer, nested ...*cobra.Command) *cobra.Command {
	rememberCmd := &cobra.Command{
		Use:           "remember <fact>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Remember a fact",
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := r.Remember(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Printf("I remembered: \"%s\"\n", summary)

			return nil
		},
	}

	rememberCmd.AddCommand(nested...)

	return rememberCmd
}
