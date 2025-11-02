package main

import (
	"io"

	"github.com/goreleaser/goreleaser-pro/v2/pkg/config"
	"go.yaml.in/yaml/v4"
)

// Parse reads a goreleaser configuration from an io.Reader and returns a config.Project.
func Parse(r io.Reader) (*config.Project, error) {
	var proj config.Project
	if err := yaml.NewDecoder(r).Decode(&proj); err != nil {
		return nil, err
	}
	return &proj, nil
}
