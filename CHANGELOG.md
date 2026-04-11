# Changelog

All notable changes to `lifi` are documented here.

## [Unreleased]

## [0.1.3] - 2026-04-11

- Replaced plain `tabwriter` output with a styled table renderer using box-drawing
  separators (`│`, `┼`, `─`), bold headers, and color-coded `yes`/`no` cells
  across all commands.
- Added a pixel-art cross-chain bridge to the root usage banner alongside the
  existing LIFI text logo.
- Improved error messages for missing `--token`, `--chain`, and `--spender`
  flags — each now includes a usage hint and example invocation.
- Added fuzzy chain resolution: prefix and substring matching with "did you mean?"
  suggestions and a link to `lifi chains`.
- `lifi approve` now validates required flags before making any API calls,
  with a full usage example in the error output.
- Rewrote `README.md`: removed all hackathon references, split Earn and Composer
  into separate sections, added per-command flag tables with required-secrets columns.
- Annotated `.env.example` with required vs optional comments.
- Added `docs/configuration.md`, `docs/earn.md`, and `docs/composer.md` for
  in-depth documentation of all commands and the config system.

## [0.1.2] - 2026-04-11

- Refreshed the terminal presentation with a larger magenta `li.fi cli`
  banner and colored doctor status labels.
- Fixed `allowance --quote-file` so it accepts wrapped JSON from
  `quote --unsigned --json`.
- Updated shell completion specs to include the current `quote`, `approve`, and
  `deposit` flags.
- Aligned docs with the current LI.FI API guidance around optional API keys and
  server-side usage.

## [0.1.1] - 2026-04-11

- Added stronger `doctor` checks for wallet/address consistency, live RPC chain
  detection, and native gas balance readiness.
- Fixed the write-path simulation bug so first-time deposits skip transaction
  simulation until approval is granted instead of failing with
  `TRANSFER_FROM_FAILED`.
- Added a release helper script, operations runbook, Makefile targets, and
  licensing/changelog metadata.

## [0.1.0] - 2026-04-11

- Initial public release of `lifi`.
- Added read workflows for chains, tokens, protocols, vault discovery,
  inspection, recommendations, portfolio lookups, and status tracking.
- Added write workflows for quote generation, approvals, dry-run deposits, real
  deposits, and portfolio verification.
- Added Homebrew packaging through the `lifi-cli` tap and GitHub Actions based
  build/release automation.
