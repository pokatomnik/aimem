package dump

import (
	"context"
	"io"

	"github.com/spf13/cobra"
)

type dumper interface {
	Dump(ctx context.Context, to io.Writer) error
}

func DumpCmd(d dumper, nested ...*cobra.Command) *cobra.Command {
	dumpCmd := &cobra.Command{
		Use:           "dump",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Dumps all facts to stdout",
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Dump(cmd.Context(), &Formatter{})
		},
	}

	dumpCmd.AddCommand(nested...)

	return dumpCmd
}
