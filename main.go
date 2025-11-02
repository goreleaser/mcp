package main

import (
	"context"
	_ "embed"
	"os"
	"sync"

	goversion "github.com/caarlos0/go-version"
	"github.com/charmbracelet/fang"
)

//nolint:gochecknoglobals
var (
	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

var versionOnce = sync.OnceValue(func() goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("goreleaser-pro", "Release engineering, simplified.", website),
		goversion.WithASCIIName(asciiArt),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
})

const website = "https://goreleaser.com/mcp"

//go:embed art.txt
var asciiArt string

func main() {
	if err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithVersion(versionOnce().String()),
		fang.WithoutCompletions(),
		fang.WithoutManpage(),
	); err != nil {
		os.Exit(1)
	}
}
