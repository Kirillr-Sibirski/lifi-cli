# lifi

CLI for LI.FI Earn and Composer.

`lifi` lets you discover vaults, inspect opportunities, generate Composer quotes, execute deposits, and verify portfolio positions from the terminal.

## What `lifi` does

`lifi` combines the two main LI.FI Earn workflows in one CLI:

- **Earn** for vault discovery, analytics, and portfolio data
- **Composer** for quote generation and transaction execution

This makes the CLI useful for:

- testing deposit flows without building a frontend
- researching vaults from the terminal
- exporting JSON for scripts, agents, and internal tools
- running small real deposits for QA, demos, and integrations
- verifying positions after execution

## Install

### Homebrew

```bash
brew tap kirillr-sibirski/lifi
brew install lifi
```

### Build from source

```bash
git clone https://github.com/Kirillr-Sibirski/defi-mullet.git
cd defi-mullet
go build -o bin/lifi ./cmd/lifi
./bin/lifi version
```

## Quick start

1. Export your LI.FI API key.
2. Optionally export a wallet private key for write commands.
3. Run `lifi doctor`.
4. Discover a vault with `lifi vaults`.
5. Inspect the vault with `lifi inspect`.
6. Generate a quote with `lifi quote`.
7. Execute a deposit with `lifi deposit`.
8. Verify the resulting position with `lifi portfolio`.

Example:

```bash
export LIFI_API_KEY=your_api_key
export LIFI_WALLET_PRIVATE_KEY=your_private_key

lifi doctor

lifi vaults \
  --chain base \
  --asset USDC \
  --transactional-only \
  --sort apy \
  --limit 10

lifi inspect 0xVaultAddress

lifi quote \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 100 \
  --from-address 0xYourWallet

lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25 \
  --wait \
  --verify-position

lifi portfolio 0xYourWallet
```

## How it works

`lifi` talks to the same LI.FI surfaces developers already use in apps:

### LI.FI Earn

- Base URL: `https://earn.li.fi`
- Used for: vault discovery, protocol metadata, APY and TVL analytics, portfolio tracking

### LI.FI Composer

- Base URL: `https://li.quest/v1`
- Used for: quote generation and deposit execution
- Important detail: for Earn deposits, the selected vault address is passed as Composer's `toToken`

### General LI.FI API

- Base URL: `https://li.quest/v1`
- Used for: token metadata, chain metadata, and transaction status

## Command reference

### `lifi doctor`

Checks that your environment is ready for read and write workflows.

What it validates:

- `earn.li.fi` reachability
- `li.quest` reachability
- `LIFI_API_KEY` presence
- API key usability against LI.FI endpoints
- wallet/private key availability for write commands
- RPC availability for configured chains

Flags:

- `--json`
- `--write-checks`
- `--chain <chain>`
- `--rpc-url <url>`

### `lifi chains`

Lists chains relevant to LI.FI Earn and Composer execution.

Flags:

- `--search <query>`
- `--evm-only`
- `--json`

### `lifi protocols`

Lists supported Earn and Composer protocols.

Flags:

- `--search <query>`
- `--supports deposit|withdraw`
- `--json`

### `lifi tokens`

Resolves tokens by symbol or address and returns metadata for quote preparation and validation.

Flags:

- `--chain <chain>`
- `--token <symbol-or-address>`
- `--tags <tag[,tag]>`
- `--json`

### `lifi vaults`

Lists depositable vaults and lets you filter by chain, token, protocol, APY, and TVL.

Default columns:

- rank
- vault name
- protocol
- chain
- deposit asset
- APY
- 30d APY
- TVL
- transactional status
- vault address

Flags:

- `--chain <chain>`
- `--asset <symbol-or-address>`
- `--protocol <name>`
- `--sort apy|apy30d|tvl|name`
- `--order asc|desc`
- `--min-tvl-usd <amount>`
- `--min-apy <percent>`
- `--transactional-only`
- `--limit <n>`
- `--json`

### `lifi inspect <vault>`

Shows full vault details.

Output includes:

- vault identity
- chain and protocol
- deposit token information
- APY breakdown
- TVL metrics
- transactional support
- raw contract addresses
- warnings for missing or partial analytics fields

Flags:

- `--json`

### `lifi recommend`

Ranks vaults for a target asset and returns the best matches based on a selected strategy.

Strategies:

- `highest-apy`
- `safest`
- `balanced`

Flags:

- `--asset <symbol-or-address>`
- `--from-chain <chain>`
- `--to-chain <chain>`
- `--strategy highest-apy|safest|balanced`
- `--min-tvl-usd <amount>`
- `--limit <n>`
- `--json`

### `lifi quote`

Generates a Composer quote for depositing into a vault.

Required inputs:

- source chain
- source token
- amount
- source wallet address
- target vault

Quote output includes:

- selected vault
- source token and amount
- target chain
- route or execution tool
- estimated output
- gas estimate
- minimum received
- approval requirement
- raw transaction target, value, and calldata

Flags:

- `--vault <address>`
- `--from-chain <chain>`
- `--to-chain <chain>`
- `--from-token <symbol-or-address>`
- `--amount <human>`
- `--amount-wei <raw>`
- `--from-address <address>`
- `--to-address <address>`
- `--slippage-bps <bps>`
- `--preset <preset>`
- `--allow-bridges <bridge[,bridge]>`
- `--deny-bridges <bridge[,bridge]>`
- `--allow-exchanges <exchange[,exchange]>`
- `--deny-exchanges <exchange[,exchange]>`
- `--json`
- `--raw`

Notes:

- `--to-chain` defaults to the vault chain
- `--vault` maps to Composer's `toToken`
- `--amount` and `--amount-wei` are mutually exclusive

### `lifi allowance`

Checks whether a wallet has enough allowance for a token and spender.

Flags:

- `--chain <chain>`
- `--token <symbol-or-address>`
- `--owner <address>`
- `--spender <address>`
- `--amount <human>`
- `--quote-file <path>`
- `--json`

### `lifi approve`

Sends an ERC-20 approval transaction for a token and spender.

Flags:

- `--chain <chain>`
- `--token <symbol-or-address>`
- `--spender <address>`
- `--amount <human|max>`
- `--yes`
- `--json`

### `lifi deposit`

Executes a full Earn deposit flow from the terminal.

Execution flow:

1. validate wallet and RPC configuration
2. resolve the target vault
3. fetch a Composer quote
4. inspect balance and allowance
5. send approval when needed
6. show a final execution summary
7. broadcast the Composer transaction
8. wait for the source transaction receipt
9. poll LI.FI status for cross-chain flows when needed
10. verify the resulting Earn position

Flags:

- `--vault <address>`
- `--from-chain <chain>`
- `--to-chain <chain>`
- `--from-token <symbol-or-address>`
- `--amount <human>`
- `--from-address <address>`
- `--to-address <address>`
- `--slippage-bps <bps>`
- `--approve auto|always|never`
- `--wait`
- `--verify-position`
- `--yes`
- `--dry-run`
- `--json`

Behavior:

- prompts before broadcasting unless `--yes` is set
- `--dry-run` stops after quote and readiness checks
- `--approve auto` only approves when necessary
- `--verify-position` checks the Earn portfolio endpoint after execution

### `lifi portfolio <address>`

Shows the current Earn positions for an address.

Default columns:

- protocol
- chain
- vault
- supplied asset
- balance
- value USD
- APY
- realized and unrealized metrics when available

Flags:

- `--chain <chain>`
- `--protocol <name>`
- `--asset <symbol-or-address>`
- `--json`

### `lifi status`

Tracks LI.FI execution state for a transaction hash.

Flags:

- `--tx-hash <hash>`
- `--from-chain <chain>`
- `--to-chain <chain>`
- `--bridge <name>`
- `--watch`
- `--interval <duration>`
- `--json`

### Utility commands

- `lifi completion <shell>`
- `lifi version`
- `lifi config init`
- `lifi config show`

## Global flags

All commands support:

- `--config <path>`
- `--profile <name>`
- `--json`
- `--verbose`
- `--quiet`
- `--no-color`

## Configuration

`lifi` reads configuration from flags, environment variables, and a local config file.

Precedence:

1. command flags
2. environment variables
3. config file defaults

### Environment variables

- `LIFI_API_KEY`
- `LIFI_WALLET_PRIVATE_KEY`
- `LIFI_WALLET_ADDRESS`
- `LIFI_DEFAULT_FROM_CHAIN`
- `LIFI_DEFAULT_SLIPPAGE_BPS`
- `LIFI_RPC_<CHAIN_KEY>`

Examples:

- `LIFI_RPC_BASE`
- `LIFI_RPC_ARBITRUM`
- `LIFI_RPC_ETHEREUM`

### Config file

Default path:

- `~/.config/lifi/config.yaml`

Example:

```yaml
profile: default
api:
  lifi_api_key: ""
defaults:
  from_chain: base
  slippage_bps: 50
  output: table
wallet:
  address: "0x..."
  private_key_env: "LIFI_WALLET_PRIVATE_KEY"
rpcs:
  base: "https://..."
  arbitrum: "https://..."
```

## Output modes

### Human mode

- colorized tables
- concise summaries
- explorer links
- warnings and next-step hints

### JSON mode

- available on all major commands
- stable field names for scripts and agents
- no extra stdout noise when `--json` is enabled

## Safety

Read-only commands:

- `doctor`
- `chains`
- `protocols`
- `tokens`
- `vaults`
- `inspect`
- `recommend`
- `quote`
- `allowance`
- `portfolio`
- `status`

Write commands:

- `approve`
- `deposit`

Safety rules:

- quote commands never broadcast
- write commands require explicit confirmation unless `--yes` is set
- execution summaries always show source chain, token, amount, vault, and recipient
- write commands fail fast when RPC or wallet configuration is missing
- vault warnings surface when a vault is not transactional or when analytics fields are missing

## Example workflows

### Discover the best Base USDC vaults

```bash
lifi vaults \
  --chain base \
  --asset USDC \
  --sort apy \
  --transactional-only \
  --limit 10
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
  --amount 100 \
  --from-address 0xYourWallet
```

### Run a deposit

```bash
lifi deposit \
  --vault 0xVaultAddress \
  --from-chain base \
  --from-token USDC \
  --amount 25 \
  --wait \
  --verify-position
```

### Watch execution status

```bash
lifi status \
  --tx-hash 0xSourceTransactionHash \
  --watch
```

### Export machine-readable output

```bash
lifi vaults --chain base --asset USDC --json
```

## Repository layout

```text
cmd/lifi/
internal/config/
internal/output/
internal/doctor/
internal/earn/
internal/lifi/
internal/wallet/
internal/evm/
internal/quote/
internal/portfolio/
internal/status/
```

## Packaging

`lifi` ships as:

- a standalone binary
- GitHub release artifacts
- a Homebrew package
- shell completions

The CLI is designed to be scriptable directly and to serve as a backend for future agent wrappers and editor skills.
