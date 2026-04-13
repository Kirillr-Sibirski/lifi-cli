# lifi-cli

`lifi` brings LI.FI Earn + Composer into the terminal. It lets builders
discover yield opportunities, inspect vaults, generate routes, execute
deposits, and verify positions without wiring a frontend first.

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

## Quick start

```bash
cp .env.example .env
lifi doctor --write-checks --chain opt
lifi vaults --chain opt --asset USDC --transactional-only --sort apy --limit 5
```

## Docs

The docs app lives in `/frontend`, and the authored markdown lives in
`/frontend/content`.

Hosted docs:

- [lifi-cli.vercel.app/docs/getting-started](https://lifi-cli.vercel.app/docs/getting-started)

Useful entry points:

- [getting-started.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/getting-started.md)
- [earn.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/earn.md)
- [composer.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/composer.md)
- [automation.md](/Users/kirillrybkov/Desktop/lifi/frontend/content/automation.md)

## Development

```bash
go test ./...

cd frontend
bun install
bun run build
```
