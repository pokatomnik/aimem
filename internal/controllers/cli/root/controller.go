package root

import "github.com/spf13/cobra"

func RootCmd(nested ...*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "aimem <command>",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "AI memory vector database client and MCP server",
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(nested...)

	return rootCmd
}
