# Banana Markets Keeper 

Off-chain agent for the Vault Router (Banana Markets) ERC-4626 diamond. It reads
live market state, decides a risk-gated target allocation, and submits
curator-permissioned transactions on a schedule. It deploys **no oracle and no
contracts** — it reads the strategies' existing on-chain views (and one public
API) and signs as the vault's curator.

Module: `github.com/vault-router-keeper` · Go.

> Build note: this repo builds with `GOWORK=off` (it is excluded from any parent
> Go workspace). Prefix go commands accordingly.

---

## Architecture

One poll loop (`internal/keeper`): **Perceive → Decide → Validate → Execute.**
The keeper owns no allocation logic — planning is mechanical; the brain decides.

```
tick:
  state  = perceive.Snapshot()      # on-chain vault state + live yield + live risk
  alloc  = brain.Decide(state)      # target bps per strategy (risk-gated water-fill)
  alloc  = validate.Check(alloc)    # re-check hard bounds; else fall back to current targets
  for action in plan(state, alloc): # SetAllocation / Rebalance / HarvestAll / GuardCheckpoint / FulfillWithdraw
      execute.Execute(action)       # sign + send as curator
```

Packages:

| Package | Role |
|---|---|
| `internal/perceive` | Builds the `VaultState` snapshot: `chain_reader` reads the diamond (totalAssets, idle, per-strategy target/assets/cap/quarantine, idle reserve, churn bound); `overlay_reader` layers the live **APY** from yield providers onto it. |
| `internal/brain` | `GauntletLiteDecider` (`gauntlet_lite.go`) + `RiskModel`/`RiskInputs` (`risk.go`). The only component with allocation intelligence. |
| `internal/risk` | Live **risk** providers per venue + `composite.go` router (StrategyID → adapter). |
| `internal/validate` | Re-checks every hard bound off-chain; on failure emits the current on-chain targets (a no-op) so an out-of-bounds allocation is never pushed. |
| `internal/trigger` | Scheduler — fires guard / harvest / rebalance on their cadences. |
| `internal/execute` | `LocalExecutor` — signs and sends curator transactions (live or dry-run). |
| `internal/chain` | RPC client wiring. |
| `internal/bindings` | abigen contract bindings (`vault`, `aavedata`, `comet`, …). |
| `internal/config` | Env parsing. |
| `cmd/keeper` | Entrypoint: wires providers from config and runs the loop. |

**Design invariants:** every number traces to a live read — an unconfigured or
failing source returns `ok=false` (degrade to the no-data path), never a
fabricated value. Risk is a hard gate; yield is the objective. Execution is
curator-permissioned, observable on-chain, and overridable by governance.

---

## The brain (risk-gated water-fill)

`GauntletLiteDecider.Decide` maximizes yield subject to risk:

1. **Per-strategy pass** — for each venue compute the scoring yield
   `effectiveYield(APY)`, the expected loss `EL`, and a **risk-adjusted cap**
   `effectiveCapBps(ownerCap, EL)`. Quarantined/killed venues → cap 0.
2. **Water-fill** — pour the investable budget (`10000 − idleReserveBps`) in
   `StepBps` (25) chunks into the highest marginal score `yield − λ·EL`, never
   exceeding a venue's effective cap.
3. **Hysteresis** — moves below `MinDeltaBps` (50) snap back to the current
   target (no churn on noise).
4. **Churn clamp** — total movement is bounded by the vault's `MaxRebalanceDelta`.

`effectiveCapBps(ownerCap, EL)`: `EL ≤ SoftThreshold (2%)` → full owner cap;
`EL ≥ KillThreshold (10%)` → 0; linear taper between. `EL = max(closed-form
PD×LGD, vendor EL)`. So governance sets the ceiling (owner cap); live risk
decides how much of it is actually used.

**Separation of powers (enforced on-chain):** the *owner* sets caps, idle
reserve, fees and the circuit breaker; the keeper holds only the *curator* role
(set allocation, rebalance, harvest) and can never exceed an owner cap.

---

## Strategies

Each venue is an isolated diamond facet on-chain and a yield+risk adapter here,
routed by `bytes32` StrategyID. Every read is keyless and live unless noted.

### Aave V3 — `bytes32("aave")`
- **Yield** (`internal/risk/aave/yield.go`): Aave V3 `ProtocolDataProvider.getReserveData().liquidityRate` (RAY, 1e27) → supply APR = `liquidityRate / 1e27` (already annualized).
- **Risk** (`internal/risk/aave/chaos.go`, `AaveChaosReader`): reads utilization (`totalDebt / totalAToken`), supply cap, and active/frozen/paused state → `LiquidityHaircut` (utilization curve) + modeled EL (ramps above 80% utilization) + a supply-cap liquidity ceiling. `aave.go` holds a conservative configured-profile fallback for when the data provider is unset.
- **Env**: `KEEPER_AAVE_DATA_PROVIDER`, `KEEPER_AAVE_ASSET`, `KEEPER_AAVE_STRATEGY_IDS`.

### Compound V3 (Comet) — `bytes32("compound")`
- **Yield** (`internal/risk/compound/yield.go`): `Comet.getSupplyRate(getUtilization())` (per-second, 1e18) → APR = `rate / 1e18 × 31_536_000`. (Per-second spot rate, unlike Aave's pre-annualized value.)
- **Risk** (`internal/risk/compound/risk.go`): `Comet.getUtilization()` (1e18) → `LiquidityHaircut` + modeled EL, mirroring the Aave utilization curve.
- **Binding**: `internal/bindings/comet` (minimal: `getUtilization`, `getSupplyRate`, `balanceOf`, `baseToken`).
- **Env**: `KEEPER_COMPOUND_COMET`, `KEEPER_COMPOUND_ASSET`, `KEEPER_COMPOUND_STRATEGY_IDS`.
- **Note**: deployed on Arbitrum One (and Ethereum mainnet); the Arbitrum-native lending venue that replaces Pendle there.

### Morpho (MetaMorpho) — `bytes32("morpho")`
- **Yield** (`internal/risk/morpho/yield.go`): net APY from Morpho's public Blue **GraphQL API** (no auth).
- **Risk** (`internal/risk/morpho/onchain.go`, `MorphoBlueReader`): **keyless on-chain** read over the Morpho Blue singleton — per-market utilization and withdrawal liquidity → EL. The free alternative to a paid rating vendor.
- **Credora (optional)**: `internal/risk/credora` — an auth-gated paid risk vendor that *overrides* the on-chain reader for its ids when an endpoint+key are supplied. Dormant by default.
- **Env**: `KEEPER_MORPHO_API`, `KEEPER_MORPHO_VAULTS`, `KEEPER_MORPHO_BLUE` (Credora: `KEEPER_CREDORA_ENDPOINT`, `KEEPER_CREDORA_API_KEY[_ENV]`, `KEEPER_CREDORA_MARKETS`).

### Pendle PT — `bytes32("pendle")`
- **Yield** (`internal/risk/pendle`): implied fixed-to-maturity APY from the canonical `PendlePYLpOracle`.
- **Risk**: a maturity/exit-liquidity gate that sets a hard `LiquidityCapBps` (capital can only hold what the async withdraw queue can cover before maturity).
- **Env**: `KEEPER_PENDLE_ORACLE`, `KEEPER_PENDLE_MARKETS` (`id=market:pt:twapSeconds`).
- **Note**: viable on Ethereum mainnet (used in the mainnet-fork demo); not deployed for the Arbitrum target — Compound takes its slot there.

---

## Configuration

Core env (see `.env.example` / `.env.fork`):

| Var | Meaning |
|---|---|
| `KEEPER_RPC_URL` | Chain RPC endpoint. |
| `KEEPER_VAULT_ADDRESS` | The diamond (vault) address. |
| `KEEPER_CHAIN_ID` | Chain id. |
| `KEEPER_PRIVATE_KEY` / `KEEPER_KEY_ENV` | Curator signing key (env-indirected). |
| `KEEPER_DRY_RUN` | `true` = decide + log, never send. |
| `KEEPER_BRAIN` | Decider selection (`gauntlet-lite`). |
| `KEEPER_RISK_PROVIDER` | `static` (closed-form floor only) or `composite` (route each id to its live adapter). |
| `KEEPER_POLL_INTERVAL` / `KEEPER_GUARD_INTERVAL` / `KEEPER_HARVEST_INTERVAL` / `KEEPER_REBALANCE_MIN_INTERVAL` | Loop cadences. |

Per-strategy vars are listed under each strategy above. An adapter whose
address/endpoint is empty is simply dormant (`ok=false`), so `composite` with no
addresses behaves identically to `static`.

---

## Build & run

```sh
GOWORK=off go build -o bin/keeper ./cmd/keeper     # or: make build
GOWORK=off go test ./...                            # or: make test
GOWORK=off go test -tags integration ./internal/integration/...   # make test-integration

# run against an env profile
set -a && . ./.env.fork && set +a
export KEEPER_VAULT_ADDRESS=0x...    # the deployed diamond
./bin/keeper
```

`scripts/fork-demo.sh` stands up a full end-to-end demo on an Ethereum mainnet
fork (anvil → real-facet diamond → keeper with the composite live risk layer →
simulated deposits/withdrawals). `internal/bindings` is regenerated with
`make bindings`.
