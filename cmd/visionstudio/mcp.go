package main

import (
	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/mcpserver"
)

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP stdio server for agent sessions",
		Long: `Run the PRISM Control MCP server over stdio.
Register it in .mcp.json for automatic agent integration:

  {
    "mcpServers": {
      "prism-build": {
        "command": "visionstudio",
        "args": ["mcp", "--dsn", "root:@tcp(127.0.0.1:3306)/visionstudio"]
      }
    }
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			return mcpserver.Run(cmd.Context(), svc)
		},
	}
}
