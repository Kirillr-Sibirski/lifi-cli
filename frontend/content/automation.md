# Automation and agents

`lifi` is designed to be scriptable through `--json`.

## Output conventions

The JSON shape depends on the command family:

- read commands like `vaults`, `inspect`, `recommend`, and `portfolio` return
  the underlying API objects directly
- `quote --json` returns the raw LI.FI quote object directly
- `quote --unsigned --json` wraps the quote in an automation-friendly payload
- `deposit --dry-run --json` and `deposit --json` return a staged execution
  payload with `stage`, `status`, `message`, `quote`, and `preflight`
- failures return a normalized error payload with `stage`, `status`, `message`,
  and `code`

This means the most automation-friendly write flows are:

```bash
lifi quote ... --unsigned --json
lifi deposit ... --dry-run --json
lifi deposit ... --wait --verify-position --json
```

## Common patterns

List vaults for ranking:

```bash
lifi vaults --chain base --asset USDC --transactional-only --json
```

Prepare a quote for downstream signing:

```bash
lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 50 \
  --unsigned \
  --json
```

That JSON can also be fed back into:

```bash
lifi allowance --quote-file /path/to/quote.json --owner 0xYourWallet --json
```

Run a safe preflight:

```bash
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 50 \
  --dry-run \
  --json
```

## JSON examples

### `lifi vaults --json`

This returns an array of Earn vault objects. The exact payload is the API
object, so expect many more fields than shown here.

```json
[
  {
    "id": "base:0xVaultAddress",
    "slug": "csusdc-morpho-v1",
    "name": "CSUSDC",
    "address": "0xVaultAddress",
    "chainId": 8453,
    "network": "Base",
    "isTransactional": true,
    "isRedeemable": true,
    "protocol": {
      "name": "morpho-v1",
      "url": "https://morpho.org"
    },
    "analytics": {
      "apy": {
        "total": 5.28,
        "base": 4.91,
        "reward": 0.37
      },
      "apy30d": 7.57,
      "tvl": {
        "usd": "4460216"
      }
    }
  }
]
```

### `lifi quote --json`

This returns the raw LI.FI quote object directly.

```json
{
  "type": "lifi",
  "id": "quote-id",
  "tool": "stargate",
  "action": {
    "fromChainId": 10,
    "toChainId": 8453,
    "fromAmount": "50000000",
    "fromAddress": "0xYourWallet",
    "toAddress": "0xYourWallet",
    "fromToken": {
      "symbol": "USDC",
      "address": "0x0b2c...",
      "decimals": 6
    },
    "toToken": {
      "symbol": "vault",
      "address": "0xVaultAddress",
      "decimals": 18
    }
  },
  "estimate": {
    "approvalAddress": "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
    "fromAmount": "50000000",
    "toAmount": "49810234912830912",
    "toAmountMin": "49561183738266752",
    "executionDuration": 86,
    "gasCosts": [
      {
        "amountUSD": "0.42",
        "limit": "281000"
      }
    ]
  },
  "transactionRequest": {
    "from": "0xYourWallet",
    "to": "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
    "data": "0x...",
    "value": "0",
    "chainId": 10,
    "gasPrice": "120000",
    "gasLimit": "281000"
  },
  "transactionId": "tx-id"
}
```

### `lifi quote --unsigned --json`

Use this when another system needs both the quote metadata and the unsigned
transaction payload in one response.

```json
{
  "quote": {
    "type": "lifi",
    "id": "quote-id"
  },
  "unsigned_transaction": {
    "from": "0xYourWallet",
    "to": "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
    "data": "0x...",
    "value": "0",
    "chainId": 10,
    "gasPrice": "120000",
    "gasLimit": "281000"
  },
  "stage": "quote",
  "status": "ok",
  "message": "unsigned transaction prepared"
}
```

### `lifi deposit --dry-run --json`

This is the safest payload to key off for automation, because it includes the
preflight checks and the quote in one place without broadcasting.

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
    "expected_received": "0.049810234912830912",
    "approval_address": "0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE",
    "approval_needed": true,
    "approval_amount": "50",
    "gas_policy": "auto",
    "native_balance": "21340988106244091",
    "native_balance_formatted": "0.021340988106244091",
    "token_balance": "150000000",
    "token_balance_formatted": "150",
    "estimated_gas_wei": "476600000000000",
    "estimated_gas_formatted": "0.0004766",
    "quote_gas_limit": 281000,
    "simulation_status": "skipped",
    "simulation_message": "skipped until approval is granted"
  }
}
```

### `lifi deposit --wait --verify-position --json`

After broadcast, the same result object is extended with transaction and
verification fields.

```json
{
  "stage": "completed",
  "status": "success",
  "message": "deposit completed",
  "tx_hash": "0xDepositTxHash",
  "approval_tx_hash": "0xApprovalTxHash",
  "receipt_status": 1,
  "approval_receipt_status": 1,
  "position_detected": true,
  "status_payload": {
    "status": "DONE",
    "substatus": "COMPLETED"
  },
  "positions": [
    {
      "protocol": "morpho-v1",
      "chain": "Optimism",
      "asset": "USDC"
    }
  ],
  "quote": {
    "type": "lifi",
    "id": "quote-id"
  },
  "preflight": {
    "wallet_address": "0xYourWallet",
    "approval_needed": true,
    "gas_policy": "auto"
  },
  "explorer_url": "https://optimistic.etherscan.io/tx/0xDepositTxHash"
}
```

### Error payload

Any command run with `--json` returns a normalized error object on failure:

```json
{
  "stage": "execution",
  "status": "error",
  "message": "insufficient balance: have 0 USDC, need 50 USDC",
  "code": 14
}
```

Exit code guide:

- `10`: config error
- `11`: input error
- `12`: API error
- `13`: RPC error
- `14`: execution error
- `15`: verification error

## Suggested automation contract

Expect these fields in machine output:

- `stage`
- `status`
- `message`
- `tx_hash`
- `approval_tx_hash`
- `position_detected`
- `preflight`
- `preflight.simulation_status`
- `preflight.simulation_message`

Practical recommendation:

- use `quote --unsigned --json` when another signer or agent needs calldata
- use `deposit --dry-run --json` for preflight gating
- use `deposit --wait --verify-position --json` for end-to-end execution records
- key off `stage`, `status`, and exit code first, then inspect nested fields
