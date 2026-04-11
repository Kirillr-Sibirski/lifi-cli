# lifi-cli

**`lifi`** is a command-line tool for [LI.FI](https://li.fi) — discover yield vaults, generate cross-chain deposit quotes, and execute on-chain transactions from your terminal.

```
lifi vaults --chain base --asset USDC --sort apy --limit 5
lifi deposit --vault 0xVault --from-chain base --from-token USDC --amount 100 --dry-run
```

---

## Install

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

---

## Setup

```bash
cp .env.example .env
```

Open `.env` and fill in what you need:

```bash
# Required for deposit and approve
LIFI_WALLET_PRIVATE_KEY=0xabc...

# Required for on-chain calls — add one per chain you use
# Key must match the chain's key from `lifi chains` (e.g. base, opt, arb, eth)
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_RPC_OPT=https://mainnet.optimism.io

# Optional — defaults
LIFI_DEFAULT_FROM_CHAIN=base
LIFI_DEFAULT_SLIPPAGE_BPS=50
```

Verify everything is wired up:

```bash
lifi doctor --write-checks --chain base
```

---

## Earn

### Find vaults

```bash
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 5
```

```
 # │ vault    │ protocol    │ chain       │ asset │  apy   │ apy30d │     tvl   │ tx  │ address
───┼──────────┼─────────────┼─────────────┼───────┼────────┼────────┼───────────┼─────┼─────────────────
 1 │ USDC     │ yo-protocol │ Base (8453) │ USDC  │ 16.46% │ 14.13% │ $27699993 │ yes │ 0x000000…8a65
 2 │ RE7USDC  │ morpho-v1   │ Base (8453) │ USDC  │  5.46% │  3.85% │  $2050694 │ yes │ 0x12afde…406e
 3 │ CSUSDC   │ morpho-v1   │ Base (8453) │ USDC  │  5.28% │  7.57% │  $4460216 │ yes │ 0x1d3b1c…4657
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--chain` | `$LIFI_DEFAULT_FROM_CHAIN` | Chain name, key, or numeric ID |
| `--asset` | | Token symbol or address |
| `--protocol` | | Protocol name (e.g. `morpho-v1`, `yo-protocol`) |
| `--sort` | `apy` | `apy`, `apy30d`, `tvl`, or `name` |
| `--order` | `desc` | `asc` or `desc` |
| `--min-apy` | | Minimum APY% filter |
| `--min-tvl-usd` | | Minimum TVL filter in USD |
| `--transactional-only` | | Only show vaults you can deposit into |
| `--limit` | `25` | Max results |

### Inspect a vault

```bash
lifi inspect 0x1d3b1cd0a0f242d598834b3f2d126dc6bd774657
```

Prints full APY breakdown (base / reward / 30-day), TVL, protocol, and deposit capability.

### Recommendations

```bash
lifi recommend --asset USDC --from-chain base --strategy highest-apy
lifi recommend --asset USDC --from-chain base --strategy safest --min-tvl-usd 5000000
```

Strategies: `highest-apy`, `safest` (weighted by TVL), `balanced` (default).

### Portfolio

```bash
lifi portfolio 0xYourWallet
lifi portfolio 0xYourWallet --chain base --protocol morpho-v1
```

Shows current Earn positions. No wallet key required — pass any address.

### Deposit

```bash
# 1. See the execution plan without broadcasting
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --dry-run

# 2. Broadcast, wait for confirmation, verify position appeared
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --wait \
  --verify-position
```

**Required:** `LIFI_WALLET_PRIVATE_KEY` and `LIFI_RPC_<CHAIN>`

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--vault` | required | Target vault address |
| `--from-chain` | `$LIFI_DEFAULT_FROM_CHAIN` | Source chain |
| `--from-token` | required | Token to deposit (e.g. `USDC`, `ETH`) |
| `--amount` | required | Human-readable amount (e.g. `100`) |
| `--to-chain` | vault's chain | For cross-chain deposits |
| `--dry-run` | | Preflight only — never broadcasts |
| `--wait` | | Wait for confirmation |
| `--verify-position` | | Check portfolio after confirmation |
| `--approve` | `auto` | `auto`, `always`, or `never` |
| `--approval-amount` | `exact` | `exact` or `infinite` |
| `--gas-policy` | `auto` | `auto`, `quote`, or `rpc` |
| `--yes` | | Skip confirmation prompt |
| `--skip-simulate` | | Bypass RPC simulation |

---

## Composer

### Quote

Get a routing quote before any broadcast:

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100
```

Export an unsigned transaction for external signing:

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --unsigned --json > quote.json
```

**Required:** `LIFI_WALLET_ADDRESS` or `LIFI_WALLET_PRIVATE_KEY` (address is derived from it).

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--vault` | required | Target vault address |
| `--from-chain` | required | Source chain |
| `--from-token` | required | Input token |
| `--amount` | required | Input amount |
| `--slippage-bps` | `50` | Slippage tolerance in basis points |
| `--unsigned` | | Print the unsigned transaction payload |
| `--allow-bridges` | | Comma-separated bridge allowlist |
| `--deny-bridges` | | Comma-separated bridge denylist |

### Check allowance

```bash
lifi allowance \
  --chain base \
  --token USDC \
  --owner 0xYourWallet \
  --spender 0xSpenderAddress \
  --amount 100
```

Or from a saved quote file:

```bash
lifi allowance --quote-file quote.json --owner 0xYourWallet
```

**Required:** `LIFI_RPC_<CHAIN>`

### Approve

```bash
lifi approve \
  --chain base \
  --token USDC \
  --spender 0xSpenderAddress \
  --amount 100
```

Use `--amount max` for an unlimited approval. `lifi deposit --approve auto` handles this automatically.

**Required:** `LIFI_WALLET_PRIVATE_KEY` and `LIFI_RPC_<CHAIN>`

### Track a transaction

```bash
lifi status --tx-hash 0xabc... --from-chain base
lifi status --tx-hash 0xabc... --watch   # polls until done
```

---

## Utility commands

### doctor

```bash
lifi doctor                      # check API connectivity and config
lifi doctor --write-checks       # also check wallet and RPC readiness
lifi doctor --write-checks --chain base
```

### chains

```bash
lifi chains                      # list all supported chains
lifi chains --search optimism    # filter by name
lifi chains --evm-only           # only EVM chains
```

Chain keys (the `key` column) are the canonical identifiers used in `--chain` flags and `LIFI_RPC_*` env vars. Common keys: `eth`, `bas`, `opt`, `arb`, `bsc`, `pol`. Full names (`base`, `ethereum`, `optimism`) also work via fuzzy matching.

### protocols

```bash
lifi protocols                        # all Earn protocols and Composer bridges
lifi protocols --search morpho        # filter by name
lifi protocols --supports deposit     # only protocols with deposit support
```

### tokens

```bash
lifi tokens --chain base --token USDC   # find the canonical USDC on Base
lifi tokens --chain eth --token ETH     # native token info
```

### config

```bash
lifi config init          # write a starter config file to ~/.config/lifi/config.yaml
lifi config show          # show resolved config for the active profile
lifi --profile prod <cmd> # use a named profile for one command
```

Config file (`~/.config/lifi/config.yaml`) supports named profiles for multiple wallets or environments:

```yaml
profile: default

profiles:
  default:
    defaults:
      from_chain: base
      slippage_bps: "50"
    wallet:
      private_key_env: LIFI_WALLET_PRIVATE_KEY
    rpcs:
      base: https://mainnet.base.org
      opt: https://mainnet.optimism.io

  staging:
    defaults:
      from_chain: optimism
    wallet:
      private_key_env: STAGING_PRIVATE_KEY
    rpcs:
      opt: https://my-private-rpc.example.com
```

---

## Global flags

These work with every command:

```bash
lifi --json <command>         # machine-readable JSON output
lifi --verbose <command>      # extra detail
lifi --no-color <command>     # disable colors (for pipes/CI)
lifi --profile <name> <cmd>   # switch config profile
lifi --config <path> <cmd>    # use a custom config file path
```

---

## Safety

- `deposit` always runs a preflight check before broadcasting
- `deposit --dry-run` never broadcasts — always safe to run
- Simulation (`--simulate`) is on by default and catches most bad transactions before they hit the chain
- Approvals default to `exact` amount — never approve more than needed
- Write commands prompt for confirmation unless `--yes` is set

---

## Troubleshooting

**`wallet private key is required`**  
Set `LIFI_WALLET_PRIVATE_KEY` in `.env` or export it in your shell.

**`unknown chain "xyz"`**  
Run `lifi chains --search xyz` to find the right key. Also try the numeric chain ID (e.g. `--chain 8453` for Base).

**`--token is required` / `--vault is required`**  
All flags use double-dash: `--token USDC`, not `-token USDC`.

**`no RPC URL configured for Base`**  
Add `LIFI_RPC_BASE=https://mainnet.base.org` to `.env`. The key after `LIFI_RPC_` must match the chain name or key (run `lifi chains` to check).

**`simulation failed`**  
Run `lifi deposit ... --dry-run --json` to see the full preflight. If `approval_needed: true`, the simulation is intentionally skipped — grant approval first.

**`position detected: no`**  
The portfolio index may take a moment to update. Retry with `lifi portfolio 0xYourWallet --chain base`.

---

## Development

```bash
go test ./...
LIFI_SMOKE=1 go test ./internal/cli -run TestLiveSmokeReadPath  # live network tests
```
