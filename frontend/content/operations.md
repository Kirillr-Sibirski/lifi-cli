# Operations and runbook

Use this document for the day-to-day operational checklist around `lifi`.

## Before a public release

1. Run `make vet` and `make test`.
2. Run `lifi doctor --write-checks --chain base` with a real local config.
3. Run a live read-path smoke test:
   `LIFI_SMOKE=1 go test ./internal/cli -run TestLiveSmokeReadPath -count=1`
4. Run at least one dry-run deposit against a real vault:
   `lifi deposit ... --dry-run --json`
5. If you want a live funds sanity check, use a dedicated low-balance wallet and
   a tiny amount on Base or Optimism.

## Release checklist

1. Commit and push `main`.
2. Tag the release: `git tag -a vX.Y.Z -m "vX.Y.Z"`.
3. Push the tag: `git push origin vX.Y.Z`.
4. Wait for the GitHub Release workflow to publish artifacts and checksums.
5. Update the Homebrew formula:
   `./scripts/update_formula.sh vX.Y.Z`
6. Commit the formula update and push `main`.
7. Verify install from a clean shell:
   `brew tap Kirillr-Sibirski/lifi-cli https://github.com/Kirillr-Sibirski/lifi-cli`
   `brew reinstall Kirillr-Sibirski/lifi-cli/lifi`
   `lifi version`

## Support triage

If a user reports a failure:

1. Ask them to run `lifi doctor --write-checks --chain <chain>`.
2. Ask for the exact command plus `--json` output when possible.
3. Check whether the failure is in:
   - config loading
   - LI.FI API response
   - RPC connectivity
   - allowance/approval
   - transaction broadcast
   - portfolio verification latency
4. Reproduce with `--dry-run` first before attempting a live transaction.

## Known operational expectations

- Base and Optimism are the best-tested real-funds paths.
- Earn indexing can lag slightly after a successful deposit, so verification may
  need a retry.
- Homebrew stable installs build from the tagged source tarball.
