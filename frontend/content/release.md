# Release and Homebrew

`lifi` ships through GitHub Releases, and the `lifi-cli` repository doubles as
the Homebrew tap.

## Release flow

1. Run the verification checklist in [operations.md](operations.md).
2. Update [CHANGELOG.md](../../CHANGELOG.md).
3. Commit and push `main`.
4. Tag a semantic version like `v0.1.1`.
5. Push the tag.
6. GitHub Actions runs GoReleaser and publishes archives plus `checksums.txt`.
7. Update the stable Homebrew formula:

   ```bash
   ./scripts/update_formula.sh v0.1.1
   ```

8. Commit the formula update and push `main`.
9. Verify the release page and install flow.

## GitHub release automation

The release workflow uses only the default `GITHUB_TOKEN` provided by GitHub
Actions. No extra release secret is required for the current setup.

Expected artifacts:

- darwin amd64 tarball
- darwin arm64 tarball
- linux amd64 tarball
- linux arm64 tarball
- `checksums.txt`

## Local snapshot

```bash
make snapshot
```

## Install path

```bash
brew tap Kirillr-Sibirski/lifi-cli https://github.com/Kirillr-Sibirski/lifi-cli
brew install lifi
```

For a `main` build:

```bash
brew install --HEAD Kirillr-Sibirski/lifi-cli/lifi
```

## Release verification

After the tag is pushed, confirm:

1. the workflow run is green
2. the GitHub release page exists for the tag
3. release artifacts are downloadable
4. `brew install lifi` resolves to the new stable version after the formula
   update lands
