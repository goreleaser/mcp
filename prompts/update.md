Let's make sure the GoReleaser configuration is in pristine condition.
Here's your tasks:

# 1. Update configuration

Let's update the GoReleaser configuration to latest.

We can use the GoReleaser's MCP `check` tool to grab the deprecation notices, and how to fix them.

If that's not enough, use the documentation resources to find out more details.
The resource paths to look at are:

- `docs://deprecations.md`
- `docs://customization/{feature name}.md`

Once you have all the information you need, resolve all the deprecations.

# 2. Add YAML schema annotation

While we are at it, let's also add the schema annotation to the top of the GoReleaser configuration file, as below:

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
```

If there's a `pro: true` in the configuration, or other indications that we're using GoReleaser Pro features, add this instead:

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema-pro.json
```
