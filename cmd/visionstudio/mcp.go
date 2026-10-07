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
  }

Remote mode: with --remote <url> (or $VISIONSTUDIO_REMOTE_URL) the same tools run
against a VisionStudio Cloud tenant instead of the local database. Store a
credential first with 'visionstudio cloud login', or pass one via
$VISIONSTUDIO_REMOTE_TOKEN:

  {
    "mcpServers": {
      "visionstudio": {
        "command": "visionstudio",
        "args": ["mcp", "--remote", "https://cloud.example.com/t/acme"]
      }
    }
  }

The cloud API currently covers initiatives and RMIs (create/get/list) plus
phase and program listing; tools that need anything else (claims, status
updates, specs, workflows, ...) return a "not supported in remote mode" error.`,
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
