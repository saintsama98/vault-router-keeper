# `abi/` — raw contract ABIs (source for code generation)

Drop each contract's `*.abi.json` here, then regenerate the Go bindings under
`internal/bindings/<name>/` with `make bindings` (requires `abigen` on PATH).

ABIs are extracted from the sibling **vault-router-diamond** Foundry `out/`
artifacts (the `.abi` field of each artifact JSON), so they match exactly the
interfaces that repo compiles against.

| ABI file | → bindings package | Source artifact |
|---|---|---|
| `vault.abi.json` | `internal/bindings/vault` | Diamond views + curator writes — see note |
| `pendle-oracle.abi.json` | `internal/bindings/pendleoracle` | `IPYLpOracle` (Pendle PT/YT/LP oracle) |
| `pendle-pt.abi.json` | `internal/bindings/pendlept` | `IPPrincipalToken` (PT: `expiry`, `isExpired`) |
| `aave-pool.abi.json` | `internal/bindings/aavepool` | `IAavePool` (supply/withdraw only — see note) |
| `aave-data.abi.json` | `internal/bindings/aavedata` | Aave V3 `IPoolDataProvider` (curated subset — see note) |

> **Note (Aave):** the diamond's `IAavePool` (`aave-pool.abi.json`) is
> intentionally minimal (supply/withdraw); it carries **no** reserve/utilization
> views. The live Aave risk feed therefore reads from the Aave V3
> **ProtocolDataProvider**, whose ABI is a **curated subset** in
> `aave-data.abi.json` — only the four getters the `risk/aave` `AaveChaosReader`
> calls, taken verbatim from the official `aave/aave-v3-core`
> `interfaces/IPoolDataProvider.sol`: `getReserveData`,
> `getReserveConfigurationData`, `getReserveCaps`, `getPaused`. The
> ProtocolDataProvider address (Arbitrum One
> `0x243Aa95cAC2a25651eda86e80bEe66114413c43b`, from `aave-dao/aave-address-book`)
> and the reserve asset (== `Vault.asset()`, USDC) are supplied via config
> (`KEEPER_AAVE_DATA_PROVIDER` / `KEEPER_AAVE_ASSET`); when unset the reader
> returns `ok=false` (closed-form floor). The legacy operator-profile
> `ConfiguredReader` is retained as a documented fallback. Likewise the Pendle
> Chaos killswitch / RedStone push ABIs are not present (RedStone *pull* has no Go
> SDK — PLAN.md §7); the `risk/pendle` adapter ships the maturity-structural
> liquidity gate plus a live fixed-to-maturity implied-APY overlay, both of which
> need only the already-present `pendle-oracle` / `pendle-pt` ABIs.

> **Note (vault):** `vault.abi.json` is a **curated** subset — only the 18
> fragments the keeper actually calls on the diamond, pulled by name from the
> `Vault`, `AllocatorFacet`, `GuardFacet`, `WithdrawQueueFacet` and `HarvestFacet`
> artifacts and merged into one ABI (the diamond exposes every facet at one
> address). Reads: `totalAssets`, `idleAssets`, `idleReserveBps`,
> `maxRebalanceDelta`, `strategies`, `targetAllocation`, `strategyTotalAssets`,
> `strategyCap`, `isQuarantined`, `paused`, `nextWithdrawRequestId`,
> `pendingWithdrawShares`, `withdrawRequest`. Curator writes: `setAllocation`,
> `rebalance`, `harvestAll`, `fulfillWithdraw`, `guardCheckpoint`. Regenerate the
> ABI from the diamond `out/` with the `sel`/`jq` extraction recorded in the
> keeper commit; add a fragment here only when the keeper needs to call it.

Keep both the ABIs here and the generated code in version control so the project
builds reproducibly without network access. ABIs are read-only inputs — never
hand-edit the generated `*.gen.go` files.
