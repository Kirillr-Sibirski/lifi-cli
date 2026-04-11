# Changelog

All notable changes to `lifi` are documented here.

## [Unreleased]

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
