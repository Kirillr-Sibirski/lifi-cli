# lifi-cli

**`lifi`** is a macOS/Linux command-line tool for the [LI.FI](https://li.fi) protocol — discover yield vaults, route cross-chain deposits, check allowances, and verify positions without building a frontend.

```
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 10
lifi deposit --vault 0xVault --from-chain base --from-token USDC --amount 100 --wait --verify-position
```

---

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [Earn — vault discovery & deposits](#earn)
- [Composer — quoting & routing](#composer)
- [Command reference](#command-reference)
- [Safety model](#safety-model)
- [Troubleshooting](#troubleshooting)
- [Development](#development)

---

## Install

### Homebrew (recommended)

```bash
brew tap Kirillr-Sibirski/lifi-cli https://github.com/Kirillr-Sibirski/lifi-cli
brew install lifi
```

To track the latest `main` build instead of the tagged release:

```bash
brew install --HEAD Kirillr-Sibirski/lifi-cli/lifi
```

### Build from source

```bash
git clone https://github.com/Kirillr-Sibirski/lifi-cli.git
cd lifi-cli
go build -o bin/lifi ./cmd/lifi
./bin/lifi version
```

---

## Quick start

```bash
# 1. Copy the environment template
cp .env.example .env
# Fill in LIFI_WALLET_PRIVATE_KEY and at least one LIFI_RPC_* entry

# 2. Verify your environment
lifi doctor --write-checks --chain base

# 3. Discover vaults
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 10

# 4. Inspect a vault
lifi inspect 0xVaultAddress

# 5. Get a quote (no broadcast)
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100

# 6. Dry-run the full deposit flow
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --dry-run

# 7. Broadcast for real
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --wait \
  --verify-position
```

---

## Configuration

`lifi` resolves settings in this priority order:

1. **Command flags** — highest priority, always win
2. **Environment variables** — exported shell vars or values from `.env`
3. **`~/.config/lifi/config.yaml`** — persistent file-based config with profile support
4. **Built-in defaults**

The `.env` file in your current working directory is loaded automatically on every run.

### Environment variables

| Variable | Required | Used by | Description |
|---|---|---|---|
| `LIFI_WALLET_PRIVATE_KEY` | Write commands | `deposit`, `approve` | Wallet private key for signing transactions. Address is derived automatically. |
| `LIFI_WALLET_ADDRESS` | Quote/portfolio | `quote`, `allowance`, `portfolio` | Explicit wallet address. Auto-derived from private key if not set. |
| `LIFI_RPC_<CHAIN>` | Write commands | `deposit`, `approve`, `allowance` | RPC endpoint for a chain. Suffix must match the chain key from `lifi chains`. |
| `LIFI_API_KEY` | Optional | all | LI.FI API key for higher rate limits. Public tier works without one. |
| `LIFI_DEFAULT_FROM_CHAIN` | Optional | all | Default chain when `--chain` is omitted. |
| `LIFI_DEFAULT_SLIPPAGE_BPS` | Optional | `quote`, `deposit` | Default slippage in basis points (50 = 0.5%). |

**RPC key format:** `LIFI_RPC_BASE`, `LIFI_RPC_OPT`, `LIFI_RPC_42161`  
The suffix is normalized (lowercased, stripped of `_`/`-`) and matched against the chain name, key, or numeric ID. Run `lifi chains` to find the exact key for any chain.

### Config file (`~/.config/lifi/config.yaml`)

The config file supports named profiles, letting you maintain separate settings for different wallets or environments.

```bash
lifi config init          # write a starter config file
lifi config show          # show resolved settings for the active profile
lifi --profile prod <cmd> # switch profile for one command
```

Example config file:

```yaml
profile: default

profiles:
  default:
    defaults:
      from_chain: base
      slippage_bps: "50"
    wallet:
      private_key_env: LIFI_WALLET_PRIVATE_KEY   # reads from env at runtime
    rpcs:
      base: https://mainnet.base.org
      opt: https://mainnet.optimism.io

  prod:
    defaults:
      from_chain: optimism
    wallet:
      address: "0xYourProdWallet"
      private_key_env: PROD_WALLET_KEY
    rpcs:
      opt: https://my-private-optimism-rpc.example.com
```

Full configuration reference → [docs/configuration.md](docs/configuration.md)

---

## Earn

The Earn commands interact with the LI.FI Earn API — discovering yield vaults and executing deposits.

Full reference → [docs/earn.md](docs/earn.md)

### `lifi vaults` — discover depositable vaults

```
lifi vaults [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--chain <name\|id>` | `$LIFI_DEFAULT_FROM_CHAIN` | Filter by chain |
| `--asset <symbol\|address>` | | Filter by underlying asset |
| `--protocol <name>` | | Filter by protocol (e.g. `morpho-v1`) |
| `--sort apy\|apy30d\|tvl\|name` | `apy` | Sort field |
| `--order asc\|desc` | `desc` | Sort direction |
| `--min-tvl-usd <amount>` | | Minimum TVL in USD |
| `--min-apy <percent>` | | Minimum APY |
| `--transactional-only` | `false` | Only show vaults that support deposits |
| `--limit <n>` | `25` | Maximum results |
| `--json` | | Machine-readable output |

**Required secrets:** none (read-only)

```bash
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 10
lifi vaults --chain ethereum --protocol morpho-v1 --min-tvl-usd 1000000 --json
```

---

### `lifi inspect` — full vault details

```
lifi inspect <vault-address|slug|name> [--json]
```

Prints APY breakdown (total / base / reward / 30-day), TVL, pack info, and deposit/redeem capability.

**Required secrets:** none

---

### `lifi recommend` — ranked vault suggestions

```
lifi recommend [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--asset <symbol\|address>` | | Target asset |
| `--from-chain <name\|id>` | | Source chain |
| `--to-chain <name\|id>` | | Vault chain (defaults to from-chain) |
| `--strategy highest-apy\|safest\|balanced` | `balanced` | Scoring strategy |
| `--min-tvl-usd <amount>` | | TVL floor |
| `--limit <n>` | `5` | Maximum results |

**Required secrets:** none

```bash
lifi recommend --asset USDC --from-chain base --strategy highest-apy
```

---

### `lifi portfolio` — view on-chain positions

```
lifi portfolio <address> [flags]
```

| Flag | Description |
|---|---|
| `--chain <name\|id>` | Filter by chain |
| `--protocol <name>` | Filter by protocol |
| `--asset <symbol\|address>` | Filter by asset |

**Required secrets:** none (the address is passed as an argument)

```bash
lifi portfolio 0xYourWallet --chain base
lifi portfolio 0xYourWallet --protocol morpho-v1 --json
```

---

### `lifi deposit` — execute a full deposit

```
lifi deposit --vault <address> --from-chain <chain> --from-token <token> --amount <amount> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--vault <address>` | **required** | Target vault |
| `--from-chain <name\|id>` | `$LIFI_DEFAULT_FROM_CHAIN` | Source chain |
| `--from-token <symbol\|address>` | **required** | Token to deposit |
| `--amount <human>` | **required** | Amount in human-readable units (e.g. `100`) |
| `--to-chain <name\|id>` | vault's chain | Override destination chain |
| `--from-address <address>` | derived from key | Override source wallet |
| `--slippage-bps <n>` | `$LIFI_DEFAULT_SLIPPAGE_BPS` | Slippage tolerance |
| `--approve auto\|always\|never` | `auto` | ERC-20 approval mode |
| `--approval-amount exact\|infinite` | `exact` | Approval sizing |
| `--gas-policy auto\|quote\|rpc` | `auto` | Gas pricing strategy |
| `--wait` | `false` | Wait for on-chain confirmation |
| `--verify-position` | `false` | Poll portfolio after confirmation |
| `--wait-timeout <duration>` | `5m` | Max wait for confirmation |
| `--portfolio-timeout <duration>` | `1m` | Max wait for position to appear |
| `--dry-run` | `false` | Prepare and preflight only, no broadcast |
| `--simulate` | `true` | RPC-simulate before broadcast |
| `--skip-simulate` | `false` | Bypass simulation |
| `--yes` | `false` | Skip confirmation prompt |

**Required secrets:**
- `LIFI_WALLET_PRIVATE_KEY` — signs the transaction
- `LIFI_RPC_<CHAIN>` — RPC for the source chain

```bash
# Safe preflight (no broadcast)
lifi deposit \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --dry-run

# Full broadcast with confirmation and position check
lifi deposit \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --wait \
  --verify-position

# Cross-chain deposit (USDC on Optimism → vault on Base)
lifi deposit \
  --vault 0xVault \
  --from-chain opt \
  --from-token USDC \
  --to-chain base \
  --amount 50 \
  --wait
```

---

## Composer

The Composer commands use the LI.FI Routing API to build, inspect, and execute cross-chain swap/bridge quotes.

Full reference → [docs/composer.md](docs/composer.md)

### `lifi quote` — generate a routing quote

```
lifi quote --vault <address> --from-chain <chain> --from-token <token> --amount <amount> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--vault <address>` | **required** | Target vault (used as the to-token) |
| `--from-chain <name\|id>` | **required** | Source chain |
| `--from-token <symbol\|address>` | **required** | Input token |
| `--amount <human>` | **required** | Input amount (or use `--amount-wei`) |
| `--amount-wei <raw>` | | Amount in base units (mutually exclusive with `--amount`) |
| `--from-address <address>` | derived from key | Source wallet |
| `--to-address <address>` | same as from | Destination wallet |
| `--slippage-bps <n>` | `$LIFI_DEFAULT_SLIPPAGE_BPS` | Slippage in basis points |
| `--allow-bridges <csv>` | | Allowlisted bridge keys |
| `--deny-bridges <csv>` | | Denylisted bridge keys |
| `--allow-exchanges <csv>` | | Allowlisted exchange keys |
| `--deny-exchanges <csv>` | | Denylisted exchange keys |
| `--unsigned` | `false` | Print the unsigned transaction payload for external signing |
| `--raw` | `false` | Print raw transaction request details |

**Required secrets:** `LIFI_WALLET_ADDRESS` (or `LIFI_WALLET_PRIVATE_KEY` to derive it)

```bash
# Standard quote
lifi quote \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100

# Get an unsigned tx for external signing
lifi quote \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --unsigned \
  --json > quote.json
```

---

### `lifi allowance` — check token approval

```
lifi allowance [--chain <chain>] [--token <token>] [--owner <address>] [--spender <address>] [--amount <human>] [--quote-file <path>]
```

Can be driven by a saved quote file (from `lifi quote --unsigned --json`):

```bash
lifi allowance --quote-file quote.json --owner 0xYourWallet
```

**Required secrets:** `LIFI_RPC_<CHAIN>` for on-chain lookup

---

### `lifi approve` — send an ERC-20 approval

```
lifi approve --chain <chain> --token <token> --spender <address> --amount <human|max> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--chain <name\|id>` | **required** | Chain to approve on |
| `--token <symbol\|address>` | **required** | Token to approve |
| `--spender <address>` | **required** | Spender contract address |
| `--amount <human\|max>` | **required** | Approval amount or `max` for unlimited |
| `--gas-policy auto\|rpc` | `auto` | Gas pricing |
| `--yes` | `false` | Skip confirmation prompt |

**Required secrets:**
- `LIFI_WALLET_PRIVATE_KEY`
- `LIFI_RPC_<CHAIN>`

```bash
lifi approve \
  --chain base \
  --token USDC \
  --spender 0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE \
  --amount 100
```

---

### `lifi status` — track a transaction

```
lifi status --tx-hash <hash> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--tx-hash <hash>` | **required** | Transaction hash |
| `--from-chain <name\|id>` | | Source chain |
| `--to-chain <name\|id>` | | Destination chain |
| `--bridge <key>` | | Bridge or tool key |
| `--watch` | `false` | Poll continuously until terminal status |
| `--interval <duration>` | `5s` | Polling interval |

**Required secrets:** none

---

## Command reference

### Read commands

| Command | Description |
|---|---|
| `lifi doctor` | Check environment, RPC reachability, and wallet readiness |
| `lifi chains` | List supported chains with their keys and IDs |
| `lifi protocols` | List Earn protocols and Composer bridges/exchanges |
| `lifi tokens` | Resolve or browse tokens by symbol or address |
| `lifi vaults` | Discover and filter yield vaults |
| `lifi inspect <vault>` | Full details for a single vault |
| `lifi recommend` | Rank vaults by strategy score |
| `lifi quote` | Build a Composer routing quote |
| `lifi allowance` | Check ERC-20 token approval |
| `lifi portfolio <address>` | View Earn positions for a wallet |
| `lifi status` | Track a cross-chain transaction |

### Write commands

| Command | Description |
|---|---|
| `lifi approve` | Send an ERC-20 approval transaction |
| `lifi deposit` | Execute a full Earn deposit (preflight + broadcast) |

### Utility commands

| Command | Description |
|---|---|
| `lifi config init` | Write a starter `config.yaml` |
| `lifi config show` | Show resolved configuration for the active profile |
| `lifi completion bash\|zsh\|fish` | Print shell completion script |
| `lifi version` | Print version |

### Global flags

These flags can be placed before or after the command name:

| Flag | Description |
|---|---|
| `--config <path>` | Path to config file |
| `--profile <name>` | Config profile to use |
| `--json` | Machine-readable JSON output |
| `--verbose` | Enable verbose output |
| `--quiet` | Suppress non-essential output |
| `--no-color` | Disable ANSI color output |

---

## Safety model

`lifi` defaults to a conservative execution model:

- `deposit` always runs a preflight check before broadcasting
- Write commands prompt for confirmation unless `--yes` is set
- `deposit --simulate` is enabled by default (RPC simulation before broadcast)
- `deposit --dry-run` never broadcasts — safe to run at any time
- Simulations are skipped automatically when an approval is still needed
- Approval mode is controlled by `--approve auto|always|never`
- Approval sizing is controlled by `--approval-amount exact|infinite`
- Gas behavior is controlled by `--gas-policy auto|quote|rpc`

**Recommended defaults:**
- Keep `--approval-amount exact` — never approve more than you need
- Keep `--gas-policy auto` — lets the CLI pick the safest estimate
- Use a dedicated low-balance wallet for testing
- Verify RPC connectivity with `lifi doctor --write-checks`

Security guidance → [docs/security.md](docs/security.md)

---

## Troubleshooting

### `wallet private key is required`

Set `LIFI_WALLET_PRIVATE_KEY` in `.env` or export it in your shell.

### `chain is required` / `unknown chain`

Pass `--chain` explicitly or set `LIFI_DEFAULT_FROM_CHAIN`. Run `lifi chains` to see valid chain names and keys.

### `--token is required`

Flags must use double-dash syntax: `--token USDC`, not `-token USDC`.

### `no RPC URL configured for <chain>`

Add `LIFI_RPC_<CHAIN_KEY>=<url>` to `.env`. The suffix must match the chain key shown by `lifi chains`. The chain's numeric ID also works (e.g. `LIFI_RPC_8453` for Base).

### `approval is required but --approve=never was set`

Use `--approve auto` or run `lifi approve` manually before the deposit.

### `simulation failed`

Check the token, chain, amount, and vault again:

```bash
lifi quote ...
lifi deposit ... --dry-run --json
```

If `approval_needed` is `true`, simulation is skipped until the approval is granted — this is expected.

### `position detected: no`

The portfolio index may not have updated yet. Wait a moment and retry:

```bash
lifi portfolio 0xYourWallet --chain base
```

---

## Development

```bash
go test ./...
```

Optional live smoke tests (requires a real network):

```bash
LIFI_SMOKE=1 go test ./internal/cli -run TestLiveSmokeReadPath
```

### Docs

- [docs/configuration.md](docs/configuration.md) — config file format, profiles, env vars
- [docs/earn.md](docs/earn.md) — Earn commands in depth
- [docs/composer.md](docs/composer.md) — Composer commands in depth
- [docs/security.md](docs/security.md) — security recommendations
- [docs/automation.md](docs/automation.md) — scripting and JSON output patterns
- [docs/supported-chains.md](docs/supported-chains.md) — chain key reference
- [docs/release.md](docs/release.md) — release and packaging process
- [docs/operations.md](docs/operations.md) — operational runbook
- [CHANGELOG.md](CHANGELOG.md)
- [LICENSE](LICENSE)
