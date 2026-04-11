# Release and Homebrew

`lifi` ships through GitHub Releases, and the `lifi-cli` repository doubles as
the Homebrew tap.

## Release flow

1. Merge to `main`.
2. Tag a semantic version like `v0.2.0`.
3. Push the tag.
4. GitHub Actions runs GoReleaser.
5. GoReleaser builds darwin/linux archives and checksums.
6. Update [`Formula/lifi.rb`](../Formula/lifi.rb) to point at the new tagged
   source tarball and checksum.
7. Push the formula update to `main`.

## Required secrets

- `GORELEASER_CURRENT_TAG`

## Local snapshot

```bash
goreleaser release --snapshot --clean
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
