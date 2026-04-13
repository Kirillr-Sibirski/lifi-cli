# Supported chains

`lifi` relies on live LI.FI chain metadata plus configured RPCs.

Practical expectations:

- EVM chains only
- best-tested paths today are Base and Optimism
- write commands require a working RPC URL for the source chain
- portfolio verification depends on LI.FI Earn portfolio indexing

Before using a chain for real funds:

```bash
lifi doctor --write-checks --chain base
```

For supported chains in your current install:

```bash
lifi chains --evm-only
```
