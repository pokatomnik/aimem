package mcp

import "github.com/spf13/cobra"

func NewMCPController(cmds ...*cobra.Command) *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:           "mcp <type>",
		Short:         "Start MCP server",
		SilenceUsage:  false,
		SilenceErrors: false,
		Example:       "mcp stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	mcpCmd.AddCommand(cmds...)

	return mcpCmd
}
