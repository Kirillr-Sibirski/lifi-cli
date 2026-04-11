# Release and Homebrew

`lifi` ships through GitHub Releases and a Homebrew tap.

## Release flow

1. Merge to `main`.
2. Tag a semantic version like `v0.2.0`.
3. Push the tag.
4. GitHub Actions runs GoReleaser.
5. GoReleaser builds darwin/linux archives and checksums.
6. The release workflow updates the `Kirillr-Sibirski/homebrew-lifi` tap.

## Required secrets

- `GORELEASER_CURRENT_TAG`
- `HOMEBREW_TAP_GITHUB_TOKEN`

## Local snapshot

```bash
goreleaser release --snapshot --clean
```

## Install path

```bash
brew tap Kirillr-Sibirski/lifi
brew install lifi
```
