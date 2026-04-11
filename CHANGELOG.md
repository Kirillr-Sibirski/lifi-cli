# Changelog

All notable changes to `lifi` are documented here.

## [Unreleased]

- Added stronger `doctor` checks for wallet/address consistency, live RPC chain
  detection, and native gas balance readiness.
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
