# lifi-cli

[![license: apache-2.0](https://img.shields.io/badge/license-apache%202.0-4b5563.svg)](https://github.com/Kirillr-Sibirski/lifi-cli/blob/main/LICENSE)
[![version: 0.1.6](https://img.shields.io/badge/version-0.1.6-f472b6.svg)](https://github.com/Kirillr-Sibirski/lifi-cli/releases)

`lifi` brings LI.FI Earn + Composer into the terminal. It lets builders
discover yield opportunities, inspect vaults, generate Composer routes,
simulate and execute deposits, and verify positions without wiring a frontend
first.

## official li.fi docs

This project wraps LI.FI's official APIs and product surfaces:

- [LI.FI Earn overview](https://docs.li.fi/earn/overview)
- [LI.FI Earn discover + deposit recipe](https://docs.li.fi/earn/recipes/discover-and-deposit)
- [LI.FI Composer overview](https://docs.li.fi/composer/overview)
- [LI.FI quote API reference](https://docs.li.fi/api-reference/get-a-quote-for-a-token-transfer)

Hosted docs are also available at:

- [lifi-cli.vercel.app/docs/getting-started](https://lifi-cli.vercel.app/docs/getting-started)

## install

```bash
brew tap Kirillr-Sibirski/lifi-cli https://github.com/Kirillr-Sibirski/lifi-cli
brew install lifi
```

Or build from source:

```bash
git clone https://github.com/Kirillr-Sibirski/lifi-cli.git
cd lifi-cli
go build -o bin/lifi ./cmd/lifi
```

## setup

```bash
cp .env.example .env
```

Open `.env` and fill in what you need:

```bash
# required for deposit, withdraw, and approve
LIFI_WALLET_PRIVATE_KEY=0xabc...

# required for on-chain calls
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_RPC_OPT=https://mainnet.optimism.io

# optional defaults
LIFI_DEFAULT_FROM_CHAIN=opt
LIFI_DEFAULT_SLIPPAGE_BPS=50
```

Verify everything is wired up:

```bash
lifi doctor --write-checks --chain opt
```

## quick start

```bash
lifi vaults --chain opt --asset USDC --transactional-only --sort apy --limit 5
lifi inspect 0x48921e52f70cec7d3425cf5103caace27fc0b2fe
lifi quote --vault 0x48921e52f70cec7d3425cf5103caace27fc0b2fe --from-chain opt --from-token USDC --amount 0.05
lifi deposit --vault 0x48921e52f70cec7d3425cf5103caace27fc0b2fe --from-chain opt --from-token USDC --amount 0.05 --dry-run
```

## earn

The Earn side of `lifi` is about discovery, inspection, recommendations,
portfolio reads, and full deposit execution.

### find vaults

```bash
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 5
```

```text
 # │ vault    │ protocol    │ chain       │ asset │  apy   │ apy30d │     tvl   │ tx  │ address
───┼──────────┼─────────────┼─────────────┼───────┼────────┼────────┼───────────┼─────┼─────────────────
 1 │ USDC     │ yo-protocol │ Base (8453) │ USDC  │ 16.46% │ 14.13% │ $27699993 │ yes │ 0x000000…8a65
 2 │ RE7USDC  │ morpho-v1   │ Base (8453) │ USDC  │  5.46% │  3.85% │  $2050694 │ yes │ 0x12afde…406e
 3 │ CSUSDC   │ morpho-v1   │ Base (8453) │ USDC  │  5.28% │  7.57% │  $4460216 │ yes │ 0x1d3b1c…4657
```

Key flags:

| Flag | Default | Description |
|---|---|---|
| `--chain` | `$LIFI_DEFAULT_FROM_CHAIN` | Chain name, key, or numeric ID |
| `--asset` | | Token symbol or address |
| `--protocol` | | Protocol name such as `morpho-v1` |
| `--sort` | `apy` | `apy`, `apy30d`, `tvl`, or `name` |
| `--order` | `desc` | `asc` or `desc` |
| `--min-apy` | | Minimum APY filter |
| `--min-tvl-usd` | | Minimum TVL filter |
| `--transactional-only` | `false` | Only show depositable vaults |
| `--limit` | `25` | Maximum results |

### inspect a vault

```bash
lifi inspect 0x1d3b1cd0a0f242d598834b3f2d126dc6bd774657
```

This prints APY breakdown, TVL, protocol metadata, chain details, and deposit
capability for a single vault.

### recommendations

```bash
lifi recommend --asset USDC --from-chain base --strategy balanced
lifi recommend --asset USDC --from-chain base --strategy safest --min-tvl-usd 5000000
lifi recommend --asset USDC --strategy highest-apy --limit 3 --json
```

Strategies:

- `highest-apy`
- `safest`
- `balanced`

### portfolio

```bash
lifi portfolio 0xYourWallet
lifi portfolio 0xYourWallet --chain base --protocol morpho-v1
```

No wallet key is required for portfolio reads. Pass any wallet address.

### deposit

```bash
# preflight only, safe to run any time
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --dry-run

# broadcast, wait for confirmation, and verify the position
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --yes \
  --wait \
  --verify-position
```

Required:

- `LIFI_WALLET_PRIVATE_KEY`
- `LIFI_RPC_<CHAIN>`

Key flags:

| Flag | Default | Description |
|---|---|---|
| `--vault` | required | Target vault address |
| `--from-chain` | `$LIFI_DEFAULT_FROM_CHAIN` | Source chain |
| `--from-token` | required | Token to deposit |
| `--amount` | required | Human-readable amount |
| `--to-chain` | vault's chain | Override destination chain |
| `--dry-run` | `false` | Preflight only, never broadcasts |
| `--wait` | `false` | Wait for confirmation |
| `--verify-position` | `false` | Poll portfolio after confirmation |
| `--approve` | `auto` | `auto`, `always`, or `never` |
| `--approval-amount` | `exact` | `exact` or `infinite` |
| `--gas-policy` | `auto` | `auto`, `quote`, or `rpc` |
| `--skip-simulate` | `false` | Bypass RPC simulation |
| `--yes` | `false` | Skip confirmation prompt |

### withdraw

```bash
# preview only
lifi withdraw \
  --vault 0xVaultAddress \
  --chain opt \
  --amount 0.049 \
  --dry-run

# broadcast
lifi withdraw \
  --vault 0xVaultAddress \
  --chain opt \
  --amount 0.049 \
  --yes
```

`--amount` is specified in vault shares. Use `lifi portfolio <address>` to see
the share balance first.

## composer

The Composer side of `lifi` handles quote generation, approvals, allowance
checks, status tracking, and the transaction payloads used by deposit flows.

### quote

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100
```

Export an unsigned transaction for downstream signing:

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --unsigned \
  --json > quote.json
```

Key flags:

| Flag | Default | Description |
|---|---|---|
| `--vault` | required | Target vault address |
| `--from-chain` | required | Source chain |
| `--from-token` | required | Input token |
| `--amount` | required | Input amount |
| `--slippage-bps` | `50` | Slippage in bps |
| `--unsigned` | `false` | Print the unsigned tx payload |
| `--allow-bridges` | | Restrict bridge choices |
| `--deny-bridges` | | Exclude bridge choices |
| `--allow-exchanges` | | Restrict exchange choices |
| `--deny-exchanges` | | Exclude exchange choices |

### allowance

```bash
lifi allowance \
  --chain base \
  --token USDC \
  --owner 0xYourWallet \
  --spender 0xSpenderAddress \
  --amount 100
```

Or drive it from a saved quote:

```bash
lifi allowance --quote-file quote.json --owner 0xYourWallet
```

### approve

```bash
lifi approve \
  --chain base \
  --token USDC \
  --spender 0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE \
  --amount 100
```

Use `--amount max` for an unlimited approval. In most cases,
`lifi deposit --approve auto` handles this for you.

### status

```bash
lifi status --tx-hash 0xabc... --from-chain opt
lifi status --tx-hash 0xabc... --from-chain opt --watch
```

## utility commands

### doctor

```bash
lifi doctor
lifi doctor --write-checks
lifi doctor --write-checks --chain base
```

### chains

```bash
lifi chains
lifi chains --search optimism
lifi chains --evm-only
```

### protocols

```bash
lifi protocols
lifi protocols --search morpho
lifi protocols --supports deposit
```

### tokens

```bash
lifi tokens --chain base --token USDC
lifi tokens --chain eth --token ETH
```

### config

```bash
lifi config init
lifi config show
lifi --profile prod config show
```

## configuration

Configuration is resolved in this order:

1. command-line flags
2. environment variables
3. `~/.config/lifi/config.yaml`
4. built-in defaults

### environment variables

| Variable | Description |
|---|---|
| `LIFI_WALLET_PRIVATE_KEY` | Required for write commands |
| `LIFI_WALLET_ADDRESS` | Optional explicit wallet address |
| `LIFI_API_KEY` | Optional LI.FI API key |
| `LIFI_DEFAULT_FROM_CHAIN` | Default chain |
| `LIFI_DEFAULT_SLIPPAGE_BPS` | Default slippage |

RPCs are chain-scoped:

```bash
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_RPC_OPT=https://mainnet.optimism.io
LIFI_RPC_8453=https://my-private-base-rpc.example.com
```

### config file

Generate a starter config:

```bash
lifi config init
```

Example:

```yaml
profile: default

profiles:
  default:
    defaults:
      from_chain: opt
      slippage_bps: "50"
    wallet:
      private_key_env: LIFI_WALLET_PRIVATE_KEY
    rpcs:
      opt: https://mainnet.optimism.io
      base: https://mainnet.base.org
```

## automation and json

`lifi` is designed to be scriptable through `--json`.

Common automation-friendly commands:

```bash
lifi vaults --chain base --asset USDC --transactional-only --json
lifi quote --vault 0xVaultAddress --from-chain base --from-token USDC --amount 50 --unsigned --json
lifi deposit --vault 0xVaultAddress --from-chain base --from-token USDC --amount 50 --dry-run --json
lifi deposit --vault 0xVaultAddress --from-chain base --from-token USDC --amount 50 --wait --verify-position --json
```

Output conventions:

- read commands such as `vaults`, `inspect`, `recommend`, and `portfolio`
  return the underlying API objects directly
- `quote --json` returns the raw LI.FI quote object directly
- `quote --unsigned --json` wraps the quote in an automation-friendly payload
- `deposit --dry-run --json` and `deposit --json` return staged execution
  payloads with `stage`, `status`, `message`, `quote`, and `preflight`
- failures return normalized error payloads with `stage`, `status`, `message`,
  and `code`

Example dry-run shape:

```json
{
  "stage": "dry-run",
  "status": "ok",
  "message": "deposit preflight completed",
  "quote": {
    "type": "lifi",
    "id": "quote-id"
  },
  "preflight": {
    "wallet_address": "0xYourWallet",
    "source_chain": "Optimism (10)",
    "source_token": "USDC",
    "source_amount": "50",
    "destination_vault": "0xVaultAddress",
    "approval_needed": true,
    "gas_policy": "auto",
    "simulation_status": "skipped"
  }
}
```

## global flags

```bash
lifi --json <command>
lifi --verbose <command>
lifi --quiet <command>
lifi --no-color <command>
lifi --profile <name> <command>
lifi --config <path> <command>
```

## supported chains

Practical expectations:

- EVM chains only
- best-tested paths today are Base and Optimism
- write commands require a working RPC URL for the source chain
- portfolio verification depends on LI.FI Earn portfolio indexing

Before using a chain with real funds:

```bash
lifi doctor --write-checks --chain base
```

## security

`lifi` can broadcast real transactions. Recommended operating model:

- use a dedicated low-balance wallet for testing
- keep `.env` only on trusted local machines
- never commit `.env`
- prefer `deposit --dry-run` before broadcasting
- keep `--approval-amount exact` unless you explicitly want an infinite approval
- run `lifi doctor --write-checks` before first use on a chain

## release and operations

Release checklist:

1. run `make vet` and `make test`
2. run `lifi doctor --write-checks --chain base`
3. run at least one dry-run deposit against a real vault
4. tag the release with `git tag -a vX.Y.Z -m "vX.Y.Z"`
5. push the tag
6. wait for the GitHub Release workflow to publish artifacts and checksums
7. update the Homebrew formula with `./scripts/update_formula.sh vX.Y.Z`
8. verify `brew install lifi`

Expected release artifacts:

- darwin amd64 tarball
- darwin arm64 tarball
- linux amd64 tarball
- linux arm64 tarball
- `checksums.txt`

## repo docs source

The deployable docs app lives in `/frontend`, and the authored markdown source
lives in `/frontend/content`.

Useful local docs entry points:

- [getting-started.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/getting-started.md)
- [earn.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/earn.md)
- [composer.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/composer.md)
- [configuration.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/configuration.md)
- [automation.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/automation.md)

## development

```bash
go test ./...

cd frontend
bun install
bun run build
```
