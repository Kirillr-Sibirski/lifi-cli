# Configuration

`lifi` resolves configuration from multiple sources, merged in this priority order (highest wins):

1. **Command-line flags** — e.g. `--chain base`, `--from-token USDC`
2. **Environment variables** — exported shell variables or values loaded from `.env`
3. **`~/.config/lifi/config.yaml`** — the persistent config file, with optional profile support
4. **Built-in defaults** — e.g. `slippage_bps: 50`

---

## Environment variables

### Wallet

| Variable | Description |
|---|---|
| `LIFI_WALLET_PRIVATE_KEY` | Hex-encoded private key (with or without `0x` prefix). Required for all write commands. The public address is derived automatically. |
| `LIFI_WALLET_ADDRESS` | Explicit wallet address. Optional when `LIFI_WALLET_PRIVATE_KEY` is set — the address is derived from the key. Set this explicitly when you want to query a wallet without holding its key. |

### RPC endpoints

```
LIFI_RPC_<CHAIN_KEY>=<url>
```

Add one entry per chain you intend to use. The `<CHAIN_KEY>` suffix is normalized (lowercased, stripped of `_`/`-`) and matched against:

- The chain's canonical key (e.g. `opt` for Optimism, `base` for Base)
- The chain's full name (e.g. `ethereum`, `arbitrumone`)
- The chain's numeric ID (e.g. `8453` for Base, `42161` for Arbitrum)

To find the correct key for any chain:

```bash
lifi chains --search optimism
```

Examples:

```bash
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_RPC_OPT=https://mainnet.optimism.io
LIFI_RPC_8453=https://my-private-base-rpc.example.com   # by chain ID
```

If no matching `LIFI_RPC_*` variable is found, `lifi` falls back to a public RPC URL from the LI.FI API (when available). This is unreliable for write operations — always configure an explicit RPC for chains you deposit on.

### API and defaults

| Variable | Default | Description |
|---|---|---|
| `LIFI_API_KEY` | _(none)_ | LI.FI API key for higher rate limits. The public tier works without a key. |
| `LIFI_DEFAULT_FROM_CHAIN` | _(none)_ | Default chain when `--chain` is omitted. |
| `LIFI_DEFAULT_SLIPPAGE_BPS` | `50` | Default slippage in basis points (100 = 1%). |

---

## `.env` file

`lifi` loads `.env` from the **current working directory** on every run. Lines are parsed as `KEY=VALUE`. Comments (`#`) and `export` prefixes are supported. Shell variables already exported take precedence — the `.env` file never overwrites them.

```bash
# .env
LIFI_WALLET_PRIVATE_KEY=0xdeadbeef...
LIFI_RPC_BASE=https://mainnet.base.org
LIFI_DEFAULT_FROM_CHAIN=base
```

> **Never commit `.env` to version control.** Add it to `.gitignore`.

---

## Config file (`config.yaml`)

The config file lives at `~/.config/lifi/config.yaml` by default. Generate a starter file with:

```bash
lifi config init
```

Override the path with `--config <path>` or inspect what is loaded with:

```bash
lifi config show
```

### Top-level structure

```yaml
# Active profile (can be overridden with --profile)
profile: default

# Top-level values apply to all profiles as a base
defaults:
  from_chain: base
  slippage_bps: "50"

wallet:
  private_key_env: LIFI_WALLET_PRIVATE_KEY   # reads this env var at runtime

rpcs:
  base: https://mainnet.base.org

# Named profiles — each one merges on top of the top-level base
profiles:
  default:
    defaults:
      from_chain: base
    wallet:
      private_key_env: LIFI_WALLET_PRIVATE_KEY
    rpcs:
      base: https://mainnet.base.org
      opt: https://mainnet.optimism.io

  prod:
    defaults:
      from_chain: optimism
    wallet:
      address: "0xYourProdWalletAddress"
      private_key_env: PROD_PRIVATE_KEY
    rpcs:
      opt: https://my-private-optimism-rpc.example.com
```

### Profile fields

| Field | Description |
|---|---|
| `defaults.from_chain` | Default chain for `--chain` |
| `defaults.slippage_bps` | Default slippage |
| `wallet.address` | Explicit wallet address |
| `wallet.private_key_env` | Name of the env var that holds the private key |
| `api.lifi_api_key` | LI.FI API key (prefer env var instead) |
| `rpcs.<key>` | RPC URL for a chain (key is the chain name or key) |

### Switching profiles

```bash
lifi --profile prod vaults --chain opt --asset USDC
lifi --profile prod deposit --vault 0xVault --from-chain opt --from-token USDC --amount 50
```

---

## `lifi config` commands

### `lifi config init`

Writes a starter `config.yaml` to the default path.

```bash
lifi config init            # fails if file already exists
lifi config init --force    # overwrite
```

### `lifi config show`

Prints the resolved configuration for the active profile — useful for debugging what secrets and RPCs are loaded.

```bash
lifi config show
lifi --profile prod config show
lifi config show --json
```

Output includes:
- Config file path and whether it exists
- `.env` path and whether it exists
- Active profile name and all available profiles
- Whether `LIFI_API_KEY`, wallet address, and wallet key are set
- Which RPC chain keys are configured
