package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var Version = "unknown"

var cmd = &cobra.Command{
	Use:               "goreleaser-mcp",
	Short:             "The GoReleaser MCP server",
	SilenceUsage:      true,
	SilenceErrors:     true,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	RunE: func(cmd *cobra.Command, _ []string) error {
		server := mcp.NewServer(&mcp.Implementation{
			Name:    "goreleaser",
			Version: Version,
		}, nil)

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

func main() {
}
