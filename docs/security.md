# Security notes

`lifi` is a power-user CLI that can broadcast real transactions.

Recommendations:

- Use a dedicated wallet for testing.
- Keep balances intentionally small.
- Prefer `.env` only on trusted local machines.
- Never commit `.env`.
- Use `deposit --dry-run` before a real broadcast.
- Keep `--approval-amount exact` unless you explicitly want an infinite approval.
- Use `lifi doctor --write-checks` before first use on a chain.
- Check that `doctor` reports a matching configured and derived wallet address.

Current secret model:

- `.env`
- exported shell environment variables
- config profiles that reference a private-key env var name
- optional `LIFI_API_KEY` stored locally for server-side or CLI use only

Do not expose `LIFI_API_KEY` in frontend code. LI.FI’s API docs explicitly warn
against client-side exposure of `x-lifi-api-key`.

OS keychain storage is not part of the current release.
