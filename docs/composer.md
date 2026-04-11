# Composer

The Composer commands wrap the LI.FI Routing API. They let you build cross-chain routing quotes, inspect ERC-20 token allowances, submit approvals, and track transaction status.

- [`lifi quote`](#lifi-quote) — generate a routing quote for a vault deposit
- [`lifi allowance`](#lifi-allowance) — check current ERC-20 approval for a spender
- [`lifi approve`](#lifi-approve) — send an ERC-20 approval transaction
- [`lifi status`](#lifi-status) — track a cross-chain transaction by hash

For the full deposit flow (quote → approve → deposit), use [`lifi deposit`](earn.md#lifi-deposit) which handles everything end-to-end.

---

## `lifi quote`

Generate a Composer routing quote for depositing into a vault. Returns the route chosen by LI.FI, expected output amounts, gas estimates, and the raw unsigned transaction payload.

```
lifi quote --vault <address> --from-chain <chain> --from-token <token> --amount <amount> [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--vault <address>` | **required** | Target vault address. Used as the `to-token` in the routing request. |
| `--from-chain <name\|id>` | **required** | Source chain. |
| `--from-token <symbol\|address>` | **required** | Input token (e.g. `USDC`, `ETH`). |
| `--amount <human>` | **required** | Input amount in human-readable units (e.g. `100`). Mutually exclusive with `--amount-wei`. |
| `--amount-wei <raw>` | | Input amount in raw base units. Mutually exclusive with `--amount`. |
| `--to-chain <name\|id>` | vault's chain | Override destination chain. Defaults to the vault's chain. |
| `--from-address <address>` | derived from key | Source wallet address. |
| `--to-address <address>` | same as from | Destination address (if different from source). |
| `--slippage-bps <n>` | `$LIFI_DEFAULT_SLIPPAGE_BPS` | Slippage tolerance in basis points. `50` = 0.5%. |
| `--preset <name>` | | LI.FI quote preset (if applicable). |
| `--allow-bridges <csv>` | | Comma-separated list of bridge keys to allow exclusively. |
| `--deny-bridges <csv>` | | Comma-separated list of bridge keys to exclude. |
| `--allow-exchanges <csv>` | | Comma-separated list of exchange keys to allow exclusively. |
| `--deny-exchanges <csv>` | | Comma-separated list of exchange keys to exclude. |
| `--unsigned` | `false` | Print the unsigned transaction payload. Useful for external signers. |
| `--raw` | `false` | Print the full raw transaction request object. |
| `--json` | | JSON output. |

### Required secrets

`LIFI_WALLET_ADDRESS` must be set (or derivable from `LIFI_WALLET_PRIVATE_KEY`) — the Composer API requires a `fromAddress` to build the transaction.

No RPC endpoint is needed for `quote` alone.

### Output

The default table shows:

| Field | Description |
|---|---|
| `tool` | Bridge or exchange used for routing |
| `from chain` | Source chain ID |
| `to chain` | Destination chain ID |
| `from token` | Input token symbol |
| `to token` | Output token symbol (the vault) |
| `from amount` | Input amount |
| `to amount` | Expected output amount |
| `min received` | Minimum output after slippage |
| `approval address` | ERC-20 spender address to approve |
| `gas usd` | Estimated gas cost |

### Examples

```bash
# Standard quote
lifi quote \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100

# Cross-chain quote: USDC on Optimism → vault on Base
lifi quote \
  --vault 0xVault \
  --from-chain opt \
  --from-token USDC \
  --to-chain base \
  --amount 50

# Export an unsigned transaction for offline/hardware wallet signing
lifi quote \
  --vault 0xVault \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --unsigned \
  --json > quote.json

# Restrict routing to specific bridges
lifi quote \
  --vault 0xVault \
  --from-chain opt \
  --from-token USDC \
  --amount 50 \
  --allow-bridges stargate

# The saved quote file can be used with lifi allowance
lifi allowance --quote-file quote.json --owner 0xYourWallet
```

---

## `lifi allowance`

Check the current ERC-20 token allowance for a wallet/spender pair. Reports whether the existing approval covers a given amount.

```
lifi allowance [--chain <chain>] [--token <token>] [--owner <address>] [--spender <address>] [--amount <human>] [--quote-file <path>] [--json]
```

### Flags

| Flag | Description |
|---|---|
| `--chain <name\|id>` | Chain to check on. |
| `--token <symbol\|address>` | Token to check. |
| `--owner <address>` | Token owner wallet address. |
| `--spender <address>` | Spender contract address. |
| `--amount <human>` | Required amount (used to determine if allowance is sufficient). |
| `--quote-file <path>` | Path to a JSON quote file from `lifi quote --unsigned --json`. Fills in chain, token, spender, and required amount automatically. |
| `--json` | JSON output. |

### Required secrets

`LIFI_RPC_<CHAIN>` — an on-chain RPC call is needed to read the allowance.

### Examples

```bash
# Manual check
lifi allowance \
  --chain base \
  --token USDC \
  --owner 0xYourWallet \
  --spender 0xSpenderAddress \
  --amount 100

# Driven from a saved quote file
lifi allowance \
  --quote-file quote.json \
  --owner 0xYourWallet

# JSON output
lifi allowance --quote-file quote.json --owner 0xYourWallet --json
```

### Output

```
 field       │ value
─────────────┼──────────────────────────────────────
 chain       │ Base (8453)
 token       │ 0xUSDCAddress
 owner       │ 0xYourWallet
 spender     │ 0xSpenderAddress
 allowance   │ 0
 required    │ 100
 sufficient  │ no
```

---

## `lifi approve`

Send an ERC-20 approval transaction. This is rarely needed manually — `lifi deposit --approve auto` handles approvals automatically. Use `lifi approve` when you want fine-grained control over the approval step.

```
lifi approve --chain <chain> --token <token> --spender <address> --amount <human|max> [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--chain <name\|id>` | **required** | Chain to approve on. |
| `--token <symbol\|address>` | **required** | Token to approve. |
| `--spender <address>` | **required** | Spender contract address (from the quote's `approval_address` field). |
| `--amount <human\|max>` | **required** | Amount to approve. Pass `max` for an unlimited (type(uint256).max) approval. |
| `--gas-policy auto\|rpc` | `auto` | Gas pricing strategy. |
| `--yes` | `false` | Skip the confirmation prompt. |
| `--json` | | JSON output. |

### Required secrets

| Secret | Purpose |
|---|---|
| `LIFI_WALLET_PRIVATE_KEY` | Signs the approval transaction |
| `LIFI_RPC_<CHAIN>` | On-chain RPC for the chain |

### Examples

```bash
# Approve exact amount
lifi approve \
  --chain base \
  --token USDC \
  --spender 0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE \
  --amount 100

# Approve unlimited (use with caution)
lifi approve \
  --chain base \
  --token USDC \
  --spender 0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE \
  --amount max \
  --yes

# Get the spender address from a quote first
SPENDER=$(lifi quote --vault 0xVault --from-chain base --from-token USDC --amount 100 --json | jq -r '.estimate.approvalAddress')
lifi approve --chain base --token USDC --spender "$SPENDER" --amount 100
```

### Find the correct chain key

Run `lifi chains` to see all supported chains and their keys. The chain key is what you pass to `--chain`:

```bash
lifi chains --search optimism
```

```
 id      │ name      │ key  │ type │ native │ relayer
─────────┼───────────┼──────┼──────┼────────┼────────
 10      │ Optimism  │ opt  │ EVM  │ ETH    │ yes
```

So `--chain opt` is correct for Optimism.

---

## `lifi status`

Track the execution status of a LI.FI cross-chain transaction by its source transaction hash.

```
lifi status --tx-hash <hash> [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--tx-hash <hash>` | **required** | Source transaction hash. |
| `--from-chain <name\|id>` | | Source chain (improves lookup accuracy). |
| `--to-chain <name\|id>` | | Destination chain. |
| `--bridge <key>` | | Bridge or tool key used for the route. |
| `--watch` | `false` | Poll continuously until the transaction reaches a terminal status. |
| `--interval <duration>` | `5s` | Polling interval when `--watch` is enabled. |
| `--json` | | JSON output (streams one object per poll when `--watch` is set). |

### Required secrets

None.

### Terminal statuses

| Status | Meaning |
|---|---|
| `DONE` | Transaction completed successfully on both source and destination. |
| `FAILED` | Transaction failed. |
| `NOT_FOUND` | Hash not recognized by LI.FI. |
| `INVALID` | Hash is malformed or on the wrong chain. |

### Examples

```bash
# One-shot status check
lifi status --tx-hash 0xabc... --from-chain base

# Watch until done
lifi status --tx-hash 0xabc... --from-chain base --to-chain opt --watch

# Stream JSON for automation
lifi status --tx-hash 0xabc... --watch --json | jq '.status'
```
