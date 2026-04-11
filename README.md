# lifi-cli

Home of the `lifi` CLI for LI.FI Earn and Composer.

`lifi` is a macOS/Linux command-line tool for discovering vaults, generating Composer quotes, running deposits, and verifying portfolio positions without building a frontend first.

## Install

### Build from source

```bash
git clone https://github.com/Kirillr-Sibirski/lifi-cli.git
cd lifi-cli
go build -o bin/lifi ./cmd/lifi
./bin/lifi version
```

### Homebrew

The `lifi-cli` repository acts as the Homebrew tap.

```bash
brew tap Kirillr-Sibirski/lifi-cli https://github.com/Kirillr-Sibirski/lifi-cli
brew install lifi
```

To install the latest `main` build instead of the tagged release:

```bash
brew install --HEAD Kirillr-Sibirski/lifi-cli/lifi
```

Release details are documented in [docs/release.md](docs/release.md).

## Quick start

1. Copy `.env.example` to `.env`.
2. Add `LIFI_API_KEY`.
3. Add RPC URLs for the chains you plan to use.
4. Optionally add `LIFI_WALLET_PRIVATE_KEY` for write commands.
5. Run `lifi doctor --write-checks`.
6. Discover a vault with `lifi vaults`.
7. Inspect it with `lifi inspect`.
8. Generate a quote with `lifi quote`.
9. Run a safe preflight with `lifi deposit --dry-run`.
10. Broadcast a real deposit with `lifi deposit --wait --verify-position`.

Example:

```bash
cp .env.example .env

lifi doctor --write-checks --chain opt

lifi vaults \
  --chain opt \
  --asset USDC \
  --transactional-only \
  --sort apy \
  --limit 5

lifi inspect 0xVaultAddress

lifi quote \
  --vault 0xVaultAddress \
  --from-chain opt \
  --from-token USDC \
  --amount 10

lifi deposit \
  --vault 0xVaultAddress \
  --from-chain opt \
  --from-token USDC \
  --amount 10 \
  --dry-run

lifi deposit \
  --vault 0xVaultAddress \
  --from-chain opt \
  --from-token USDC \
  --amount 10 \
  --wait \
  --verify-position
```

## Configuration and secrets

`lifi` reads configuration in this order:

1. command flags
2. environment variables
3. config file defaults

Supported inputs:

- project-root `.env`
- exported shell variables
- `~/.config/lifi/config.yaml`
- profile selection via `--profile`

Primary variables:

- `LIFI_API_KEY`
- `LIFI_WALLET_PRIVATE_KEY`
- `LIFI_WALLET_ADDRESS`
- `LIFI_DEFAULT_FROM_CHAIN`
- `LIFI_DEFAULT_SLIPPAGE_BPS`
- `LIFI_RPC_<CHAIN_KEY>`

Example:

```bash
LIFI_API_KEY=...
LIFI_WALLET_PRIVATE_KEY=...
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_RPC_OPT=https://mainnet.optimism.io
```

Profile-aware config example:

```yaml
profile: default
profiles:
  default:
    defaults:
      from_chain: base
      slippage_bps: "50"
    wallet:
      private_key_env: "LIFI_WALLET_PRIVATE_KEY"
    rpcs:
      base: "https://mainnet.base.org"
  prod:
    defaults:
      from_chain: optimism
    rpcs:
      optimism: "https://mainnet.optimism.io"
```

## Core workflows

### Discover vaults

```bash
lifi vaults --chain base --asset USDC --transactional-only --limit 10
```

### Inspect a vault

```bash
lifi inspect 0xVaultAddress
```

### Generate a quote

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25
```

### Export an unsigned transaction

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25 \
  --unsigned \
  --json
```

### Run a dry-run deposit

```bash
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25 \
  --dry-run
```

### Broadcast a deposit

```bash
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25 \
  --gas-policy auto \
  --approval-amount exact \
  --wait \
  --verify-position
```

### Use JSON output for automation

```bash
lifi vaults --chain base --asset USDC --json
lifi quote --vault 0xVaultAddress --from-chain base --from-token USDC --amount 25 --json
lifi deposit --vault 0xVaultAddress --from-chain base --from-token USDC --amount 25 --dry-run --json
```

## Command reference

Read-oriented commands:

- `lifi doctor`
- `lifi chains`
- `lifi protocols`
- `lifi tokens`
- `lifi vaults`
- `lifi inspect`
- `lifi recommend`
- `lifi quote`
- `lifi allowance`
- `lifi portfolio`
- `lifi status`

Write-oriented commands:

- `lifi approve`
- `lifi deposit`

Utility commands:

- `lifi config init`
- `lifi config show`
- `lifi completion <bash|zsh|fish>`
- `lifi version`

Global flags:

- `--config <path>`
- `--profile <name>`
- `--json`
- `--verbose`
- `--quiet`
- `--no-color`

Global flags can be placed before or after the command name.

## Safety model

`lifi` is designed for builders and automation users, not casual wallet users.

Safety defaults:

- `deposit` always runs a preflight before broadcast
- write commands prompt unless `--yes` is set
- `deposit --simulate` is enabled by default
- `deposit --dry-run` never broadcasts
- approvals are controlled by `--approve auto|always|never`
- approval sizing is controlled by `--approval-amount exact|infinite`
- gas behavior is controlled by `--gas-policy auto|quote|rpc`

Recommended defaults:

- keep `--approval-amount exact`
- keep `--gas-policy auto`
- use a dedicated low-balance wallet for testing
- verify RPC URLs with `lifi doctor --write-checks`

Security guidance lives in [docs/security.md](docs/security.md).

## Troubleshooting

### `wallet private key is required`

Set `LIFI_WALLET_PRIVATE_KEY` in `.env` or export it in your shell.

### `no RPC URL configured`

Add `LIFI_RPC_<CHAIN_KEY>` to `.env` or configure the chain in `config.yaml`.

### `approval is required but --approve=never was set`

Use `--approve auto` or submit a manual approval with `lifi approve`.

### `simulation failed`

Check the token, chain, amount, and vault again with:

```bash
lifi quote ...
lifi deposit ... --dry-run --json
```

### `position detected: no`

The transaction may have succeeded before the portfolio index updated. Retry:

```bash
lifi portfolio 0xYourWallet --chain base
```

## Development

Run the local test suite:

```bash
go test ./...
```

Optional live smoke tests:

```bash
LIFI_SMOKE=1 go test ./internal/cli -run TestLiveSmokeReadPath
```

Release and packaging guidance:

- [docs/release.md](docs/release.md)
- [docs/security.md](docs/security.md)
- [docs/automation.md](docs/automation.md)
- [docs/supported-chains.md](docs/supported-chains.md)
