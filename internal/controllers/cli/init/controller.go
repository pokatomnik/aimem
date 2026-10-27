package init

import (
	"github.com/pokatomnik/gomem/internal/shared/colors"
	"github.com/spf13/cobra"
)

type dumper interface {
	Dump() (bool, error)
}

func GetInitCmd(d dumper, nested ...*cobra.Command) *cobra.Command {
	initCmd := &cobra.Command{
		Use:           "init",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Initializes a configuration file. Proceed with set up manually.",
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			initialized, err := d.Dump()
			if err == nil {
				if initialized {
					colors.Primary().Println("Config initialized")
				}
			}

			return err
		},
	}

	initCmd.AddCommand(nested...)

	return initCmd
}
