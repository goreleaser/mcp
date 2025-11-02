package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

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
			Version: versionOnce().GitVersion,
		}, nil)

		server.AddPrompt(&mcp.Prompt{
			Name:  "update_config",
			Title: "Update GoReleaser Configuration",
		}, func(ctx context.Context, gpr *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{
						Content: &mcp.TextContent{Text: updatePrompt},
						Role:    mcp.Role("user"),
					},
				},
			}, nil
		})

		// if err := fs.WalkDir(docs.FS, ".", func(path string, d fs.DirEntry, err error) error {
		// 	if d.IsDir() || !strings.HasSuffix(path, ".md") {
		// 		return err
		// 	}
		//
		// 	server.AddResource(&mcp.Resource{
		// 		Meta:        mcp.Meta{},
		// 		Annotations: &mcp.Annotations{},
		// 		Description: "",
		// 		MIMEType:    "text/markdown",
		// 		Name:        path,
		// 		URI:         "docs://" + path,
		// 	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// 		bts, err := fs.ReadFile(docs.FS, path)
		// 		if err != nil {
		// 			return nil, mcp.ResourceNotFoundError(req.Params.URI)
		// 		}
		// 		return &mcp.ReadResourceResult{
		// 			Contents: []*mcp.ResourceContents{{
		// 				URI:  req.Params.URI,
		// 				Text: string(bts),
		// 			}},
		// 		}, nil
		// 	})
		//
		// 	return nil
		// }); err != nil {
		// 	return err
		// }

		mcp.AddTool(server, &mcp.Tool{
			Name:        "check",
			Description: "Checks a GoReleaser configuration for errors or deprecations",
		}, checkTool)

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}
