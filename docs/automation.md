# Automation and agents

`lifi` is designed to be scriptable through `--json`.

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
