package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/goreleaser/goreleaser-mcp/internal/yaml"
	"github.com/goreleaser/goreleaser-pro/v2/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const updatePrompt = `Let's update the goreleaser configuration to latest.

We can use the goreleaser check command to grab the deprecation notices and how to fix them.

If that's not enough, use the documentation resources to find out more details.
The resource paths to look at are:

- docs://deprecataions.md
- docs://customization/{feature name}.md
- docs://old-deprecataions.md (this one only if updating between goreleaser major versions)
`

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
		Message string `json:"message"`
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

func checkTool(ctx context.Context, _ *mcp.CallToolRequest, args checkArgs) (*mcp.CallToolResult, checkOutput, error) {
	name, bts, err := openConfig(args.Configuration)
	if err != nil {
		return nil, checkOutput{}, fmt.Errorf("could not check configuration: %w", err)
	}

	var cfg config.Project
	if err := yaml.UnmarshalStrict(bts, &cfg); err != nil {
		return nil, checkOutput{}, fmt.Errorf("invalid configuration file: %s: %w", name, err)
	}

	deprecations := findDeprecated(cfg)
	if len(deprecations) == 0 {
		return nil, checkOutput{
			Message: fmt.Sprintf("Configuration at %q is valid!", name),
		}, nil
	}

	var sb strings.Builder
	sb.WriteString("Configuration is valid, but uses the following deprecated properties:\n")
	for _, key := range slices.Collect(maps.Keys(deprecations)) {
		sb.WriteString(fmt.Sprintf("## %s\n\nInstructions: %s\n\n", key, instructions[key]))
	}
	return nil, checkOutput{
		Message: sb.String(),
	}, nil
}
