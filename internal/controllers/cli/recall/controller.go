package recall

import (
	"context"
	"fmt"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/shared/colors"
	"github.com/spf13/cobra"
)

type recaller interface {
	Recall(ctx context.Context, query string) ([]fact.FactListView, error)
}

func RecallCmd(r recaller, nested ...*cobra.Command) *cobra.Command {
	recallCmd := &cobra.Command{
		Use:           "recall <question>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Recall some facts ranged by similarity",
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			facts, err := r.Recall(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			colors.Primary().Println("Here is some related info:")
			for _, fact := range facts {
				fmt.Printf(
					"%s: %s\n",
					colors.Secondary().Sprintf("[%s]", fact.ID),
					fact.Summary,
				)
			}

			return nil
		},
	}

	recallCmd.AddCommand(nested...)

	return recallCmd
}
