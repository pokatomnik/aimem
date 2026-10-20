package getfact

import (
	"context"
	"fmt"

	"github.com/pokatomnik/gomem/internal/entities/fact"
	"github.com/pokatomnik/gomem/internal/shared/colors"
	"github.com/spf13/cobra"
)

type getFact interface {
	GetFact(ctx context.Context, id string) (fact.FactFullView, error)
}

func GetFactCmd(g getFact, nested ...*cobra.Command) *cobra.Command {
	getFactCmd := &cobra.Command{
		Use:           "fact <factID>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Get fact by ID",
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fact, err := g.GetFact(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			colors.Primary().Println("Here is a fact")
			fmt.Printf("%s: %s\n", colors.Secondary().Sprint("[Summary]"), fact.Summary)
			fmt.Printf("%s: %s\n", colors.Secondary().Sprint("[ Fact  ]"), fact.Content)

			return nil
		},
	}

	getFactCmd.AddCommand(nested...)

	return getFactCmd
}
