package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/goreleaser/goreleaser-mcp/internal/yaml"
	"github.com/goreleaser/goreleaser-pro/v2/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

//go:embed prompts/update.md
var updatePrompt string

//go:embed docs
var docs embed.FS

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
			Title: "Updates your GoReleaser configuration, getting rid of deprecations",
		}, func(context.Context, *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{
						Content: &mcp.TextContent{Text: updatePrompt},
						Role:    mcp.Role("user"),
					},
				},
			}, nil
		})

		mcp.AddTool(server, &mcp.Tool{
			Name:        "check",
			Description: "Checks a GoReleaser configuration for errors or deprecations",
		}, checkTool)

		fsys, err := fs.Sub(docs, "docs")
		if err != nil {
			return err
		}
		if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return err
			}

			server.AddResource(&mcp.Resource{
				Meta:        mcp.Meta{},
				Annotations: &mcp.Annotations{},
				Description: "",
				MIMEType:    "text/markdown",
				Name:        path,
				URI:         "docs://" + path,
			}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				bts, err := fs.ReadFile(fsys, path)
				if err != nil {
					return nil, mcp.ResourceNotFoundError(req.Params.URI)
				}
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{{
						URI:  req.Params.URI,
						Text: string(bts),
					}},
				}, nil
			})

			return nil
		}); err != nil {
			return err
		}

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

var instructions = map[string]string{
	"archives.builds":                  "replace `builds` with `ids`",
	"archives.format":                  "replace `format` with `formats` and make its value an array",
	"archives.format_overrides.format": "replace `format` with `formats` and make its value an array",
	"builds.gobinary":                  "rename `gobinary` to `tool`",
	"homebrew_casks.manpage":           "replace `manpage` with `manpages`, and make its value an array",
	"homebrew_casks.binary":            "replace `binary` with `binaries`, and make its value an array",
	"homebrew_casks.conflicts.formula": "remove the `formula: <name>` from the `conflicts` list",
	"kos.repository":                   "replace `repository` with `repositories`, and make its value an array",
	"kos.sbom":                         "the value of `sbom` can only be `spdx` or `none`, set it to `spdx` if there's any other value there",
	"nfpms.builds":                     "rename `builds` to `ids`",
	"nightly.name_template":            "rename `name_template` to `version_template`",
	"snaps.builds":                     "rename `builds` to `ids`",
	"snapshot.name_template":           "rename `name_template` to `version_template`",
}

type (
	checkArgs struct {
		Configuration string `json:"configuration,omitempty" jsonschema:"Path to the goreleaser YAML configuration file. If empty will use the default."`
	}
	checkOutput struct {
		Message      string `json:"message"`
		Filepath     string `json:"filepath"`
		Instructions string `json:"instructions,omitempty"`
	}
)

// openConfig either opens the given name (if not empty), or tries to open all
// the default names.
//
// The caller is responsible for closing the returned file.
func openConfig(name string) (string, []byte, error) {
	if name != "" {
		bts, err := os.ReadFile(name)
		return name, bts, err
	}

	for _, name := range [6]string{
		".config/goreleaser.yml",
		".config/goreleaser.yaml",
		".goreleaser.yml",
		".goreleaser.yaml",
		"goreleaser.yml",
		"goreleaser.yaml",
	} {
		bts, err := os.ReadFile(name)
		if err == nil {
			return name, bts, err
		}
	}

	return "", nil, fmt.Errorf("could not find any configuration file")
}

func checkTool(ctx context.Context, req *mcp.CallToolRequest, args checkArgs) (*mcp.CallToolResult, checkOutput, error) {
	name, bts, err := openConfig(args.Configuration)
	if err != nil {
		return nil, checkOutput{}, fmt.Errorf("could not check configuration: %w", err)
	}

	_ = req.Session.Log(ctx, &mcp.LoggingMessageParams{
		Data:  fmt.Sprintf("using configuration file at %s", name),
		Level: mcp.LoggingLevel("info"),
	})

	var cfg config.Project
	if err := yaml.UnmarshalStrict(bts, &cfg); err != nil {
		return nil, checkOutput{}, fmt.Errorf("invalid configuration file: %s: %w", name, err)
	}

	deprecations := findDeprecated(cfg)
	if len(deprecations) == 0 {
		return nil, checkOutput{
			Message:  "Configuration is valid!",
			Filepath: name,
		}, nil
	}

	res, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
		Message:         "You have deprecated configuration options in your GoReleaser config. Do you want to fix it?",
		RequestedSchema: nil,
	})
	if err != nil || res.Action == "decline" {
		return nil, checkOutput{
			Message:  "Configuration is valid, but uses deprecated options",
			Filepath: name,
		}, nil
	}

	// if action is 'cancel' let's just add the instructions anyway...
	var sb strings.Builder
	sb.WriteString("# Deprecated Options\n\n")
	sb.WriteString("Here's the instructions to fix each of deprecation:\n\n")
	for _, key := range slices.Collect(maps.Keys(deprecations)) {
		sb.WriteString(fmt.Sprintf("## %s\n\nInstructions: %s\n\n", key, instructions[key]))
	}
	return nil, checkOutput{
		Message:      "Configuration is valid, but uses deprecated options",
		Instructions: sb.String(),
		Filepath:     name,
	}, nil
}
