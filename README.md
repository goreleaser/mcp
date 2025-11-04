<p align="center">
  <img alt="GoReleaser Logo" src="https://avatars2.githubusercontent.com/u/24697112?v=3&s=200" height="200" />
  <h3 align="center">GoReleaser MCP</h3>
  <p align="center">Model Context Protocol server for GoReleaser.</p>
</p>

---

A Model Context Protocol (MCP) server for GoReleaser.
Provides tools and documentation to help AI assistants understand and work with
GoReleaser configurations.

---

## Features

- **Configuration Validation**: Check GoReleaser configurations for errors and deprecated options
- **Deprecation Fixes**: Get prompted to fix deprecated configuration options with detailed instructions
- **Embedded Documentation**: Access GoReleaser documentation directly through the MCP server
- **Update Prompt**: Use the `update_config` prompt to modernize your GoReleaser configuration

## Installation

```bash
go install github.com/goreleaser/goreleaser-mcp@latest
npm -g @goreleaser/mcp
```

## Usage

Run the MCP server:

```bash
goreleaser-mcp
```

The server communicates over stdio and implements the Model Context Protocol.

## Badges

[![Release](https://img.shields.io/github/release/goreleaser/goreleaser-mcp.svg?style=for-the-badge)](https://github.com/goreleaser/goreleaser-mcp/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/actions/workflow/status/goreleaser/goreleaser-mcp/build.yml?style=for-the-badge&branch=main)](https://github.com/goreleaser/goreleaser-mcp/actions?workflow=build)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/goreleaser/goreleaser-mcp)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![GoReportCard](https://goreportcard.com/badge/github.com/goreleaser/goreleaser-mcp?style=for-the-badge)](https://goreportcard.com/report/github.com/goreleaser/goreleaser-mcp)
