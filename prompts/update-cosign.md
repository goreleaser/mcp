Your task is to update cosign to version 3.
The process consists of:

# Updates to .goreleaser.yml

Find all the GoReleaser configuration file that have a `signs` block in them, and apply the following diff:

```diff
signs:
  - cmd: cosign
-   certificate: "${artifact}.pem"
+   signature: "${artifact}.sigstore.json"
    args:
      - sign-blob
-     - "--output-certificate=${certificate}"
-     - "--output-signature=${signature}"
+     - "--bundle=${signature}"
      - "${artifact}"
```

Basically, replace the `certificate` option with `signature: "${artifact}.sigstore.json"`, and remove both `--output-certificate` and `--output-signature` in favor of `--bundle=${signature}`.
Make sure that the `cmd` is `cosign` and the first `args` is `sign-blob` before doing any changes.

# Update to GitHub Workflows

Find the release workflow in `.github/workflows`, make sure that `cosign-installer` action is up-to-date (currently `v4`), and if it specifies an old cosign version, remove it.
For instance, if it looks like this:

```yaml
- uses: sigstore/cosign-installer@faadad0cce49287aee09b3a48701e75088a2c6ad # v4.0.0
  with:
    cosign-release: "v2.6.1"
```

Remove the `with` block. If the version specified is v3+, leave as is, but let the user know.
