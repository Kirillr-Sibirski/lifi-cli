# Earn

The Earn commands wrap the LI.FI Earn API. They let you discover yield-bearing vaults across protocols and chains, inspect their details, get recommendations, view portfolio positions, and execute on-chain deposits.

**Read commands** (no wallet needed):
- [`lifi vaults`](#lifi-vaults) — discover and filter vaults
- [`lifi inspect`](#lifi-inspect) — full details for a single vault
- [`lifi recommend`](#lifi-recommend) — ranked vault suggestions
- [`lifi portfolio`](#lifi-portfolio) — on-chain positions for a wallet

**Write commands** (require `LIFI_WALLET_PRIVATE_KEY` + `LIFI_RPC_<CHAIN>`):
- [`lifi deposit`](#lifi-deposit) — execute a full deposit

---

## `lifi vaults`

List and filter depositable yield vaults.

```
lifi vaults [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--chain <name\|id>` | `$LIFI_DEFAULT_FROM_CHAIN` | Filter by chain. Accepts chain name, key, or numeric ID. |
| `--asset <symbol\|address>` | | Filter by underlying asset symbol (e.g. `USDC`) or token address. |
| `--protocol <name>` | | Filter by protocol name (e.g. `morpho-v1`, `pendle`). |
| `--sort apy\|apy30d\|tvl\|name` | `apy` | Sort field. |
| `--order asc\|desc` | `desc` | Sort direction. |
| `--min-tvl-usd <amount>` | | Minimum TVL filter in USD (e.g. `1000000` for $1M+). |
| `--min-apy <percent>` | | Minimum APY filter (e.g. `10` for 10%+). |
| `--transactional-only` | `false` | Only show vaults that support on-chain deposits via Composer. |
| `--limit <n>` | `25` | Maximum number of results. |
| `--json` | | Output as JSON. |

### Required secrets

None — this command only calls the read-only Earn API.

### Examples

```bash
# Top 10 USDC vaults on Base by APY
lifi vaults --chain base --asset USDC --transactional-only --sort apy --limit 10

# All morpho-v1 vaults on Ethereum with $1M+ TVL
lifi vaults --chain ethereum --protocol morpho-v1 --min-tvl-usd 1000000

# Vaults with 15%+ APY, sorted by TVL descending
lifi vaults --min-apy 15 --sort tvl

# JSON output for scripting
lifi vaults --chain base --asset USDC --json | jq '.[].analytics.apy.total'
```

---

## `lifi inspect`

Print full details for a single vault — APY breakdown, TVL, protocol info, supported deposit/redeem packs, and data freshness.

```
lifi inspect <vault> [--json]
```

`<vault>` can be:
- A vault address (`0xabc...`)
- A vault slug (e.g. `csusdc-morpho-v1`)
- A vault name (e.g. `CSUSDC`)

### Required secrets

None.

### Examples

```bash
lifi inspect 0x9b5e92fd7c2ef79fddd33c6c7a3c3e5abb...
lifi inspect csusdc-morpho-v1
lifi inspect CSUSDC --json
```

---

## `lifi recommend`

Rank vaults for a target asset and chain using a scoring strategy.

```
lifi recommend [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--asset <symbol\|address>` | | Target asset to earn yield on. |
| `--from-chain <name\|id>` | | Source chain (used to filter vaults when `--to-chain` is not set). |
| `--to-chain <name\|id>` | `--from-chain` | Chain where the vault should be. |
| `--strategy highest-apy\|safest\|balanced` | `balanced` | Scoring strategy. See below. |
| `--min-tvl-usd <amount>` | | TVL floor filter. |
| `--limit <n>` | `5` | Maximum results. |
| `--json` | | JSON output. |

### Scoring strategies

| Strategy | Description |
|---|---|
| `highest-apy` | Ranked purely by total APY. Favours high-yield vaults regardless of size. |
| `safest` | Weighted towards TVL (log scale) plus base APY only. Ignores volatile reward APY. |
| `balanced` | APY + log-TVL weight (default). Balances yield and protocol maturity. |

### Required secrets

None.

### Examples

```bash
# Best USDC vault on Base by balanced score
lifi recommend --asset USDC --from-chain base

# Safest USDC vault with $500k+ TVL
lifi recommend --asset USDC --from-chain ethereum --strategy safest --min-tvl-usd 500000

# Highest-APY vault, any chain
lifi recommend --asset USDC --strategy highest-apy --limit 3 --json
```

---

## `lifi portfolio`

Show current Earn positions for a wallet address.

```
lifi portfolio <address> [flags]
```

### Flags

| Flag | Description |
|---|---|
| `--chain <name\|id>` | Filter by chain. |
| `--protocol <name>` | Filter by protocol. |
| `--asset <symbol\|address>` | Filter by asset. |
| `--json` | JSON output. |

### Required secrets

None — pass any wallet address as an argument.

### Examples

```bash
lifi portfolio 0xYourWallet
lifi portfolio 0xYourWallet --chain base
lifi portfolio 0xYourWallet --protocol morpho-v1 --json
```

---

## `lifi deposit`

Execute a full Earn deposit using the Composer routing API. The flow is:

1. Fetch a routing quote from the Composer API
2. Run a preflight check (simulation, allowance check, balance check)
3. Prompt for confirmation (skipped with `--yes` or `--dry-run`)
4. Submit the ERC-20 approval if needed
5. Broadcast the deposit transaction
6. Optionally wait for confirmation and verify the portfolio position

```
lifi deposit --vault <address> --from-chain <chain> --from-token <token> --amount <amount> [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--vault <address>` | **required** | Target vault address. |
| `--from-chain <name\|id>` | `$LIFI_DEFAULT_FROM_CHAIN` | Source chain. |
| `--from-token <symbol\|address>` | **required** | Token to deposit. |
| `--amount <human>` | **required** | Amount in human-readable units (e.g. `100` for 100 USDC). |
| `--to-chain <name\|id>` | vault's chain | Destination chain (for cross-chain deposits). |
| `--from-address <address>` | derived from key | Override source wallet address. |
| `--to-address <address>` | same as from | Override destination wallet address. |
| `--slippage-bps <n>` | `$LIFI_DEFAULT_SLIPPAGE_BPS` | Slippage tolerance in basis points. |
| `--approve auto\|always\|never` | `auto` | Approval mode. `auto` submits if needed, `never` aborts if needed. |
| `--approval-amount exact\|infinite` | `exact` | Approval sizing. `exact` approves only the required amount. |
| `--gas-policy auto\|quote\|rpc` | `auto` | Gas pricing. `auto` picks the safer of the quote and RPC estimates. |
| `--wait` | `false` | Wait for on-chain source transaction confirmation. |
| `--verify-position` | `false` | Poll portfolio after confirmation to verify position appeared. |
| `--wait-timeout <duration>` | `5m` | Maximum time to wait for confirmation. |
| `--portfolio-timeout <duration>` | `1m` | Maximum time to poll for portfolio update. |
| `--dry-run` | `false` | Preflight only — never broadcasts. Safe to use at any time. |
| `--simulate` | `true` | RPC-simulate the transaction before broadcast. |
| `--skip-simulate` | `false` | Bypass simulation (e.g. when the chain doesn't support `eth_call` simulation). |
| `--yes` | `false` | Skip the confirmation prompt. |
| `--json` | | JSON output for all stages. |

### Required secrets

| Secret | Purpose |
|---|---|
| `LIFI_WALLET_PRIVATE_KEY` | Signs the deposit (and approval if needed) transaction |
| `LIFI_RPC_<CHAIN>` | On-chain RPC for the source chain |

### Examples

```bash
# Preflight only (safe, no broadcast)
lifi deposit \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --dry-run

# Standard deposit — waits for confirmation, then checks portfolio
lifi deposit \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --wait \
  --verify-position

# Cross-chain deposit: USDC on Optimism → vault on Base
lifi deposit \
  --vault 0xVault \
  --from-chain opt \
  --from-token USDC \
  --to-chain base \
  --amount 50 \
  --wait

# Automation-friendly (no prompts, JSON output)
lifi deposit \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --yes \
  --wait \
  --verify-position \
  --json
```

### Preflight output

Before any broadcast, `lifi deposit` prints an **Execution Plan** table that shows:

| Field | Description |
|---|---|
| `wallet` | The signing wallet address |
| `source chain` | Chain the tokens come from |
| `source token` | Token being deposited |
| `source amount` | Amount in human-readable units |
| `vault` | Target vault address |
| `expected received` | Estimated vault tokens received |
| `approval address` | Spender address for the ERC-20 approval |
| `approval needed` | Whether an approval transaction is required |
| `approval amount` | Approval amount that will be submitted |
| `gas policy` | Gas pricing strategy in effect |
| `token balance` | Current token balance in the wallet |
| `native balance` | Current native token balance (for gas) |
| `estimated gas cost` | Estimated gas cost in USD |

Use `--dry-run` to see this output without broadcasting.
