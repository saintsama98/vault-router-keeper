# Vault Router Keeper — Build Plan: the Risk Brain

**Source of truth:** `/mnt/adiii_dev/Ethereum-dev/vault-router-keeper-research/risk-brain-architecture.md` (in the sibling research project — read it first; this plan implements it).
**Status:** ready to execute · **Target model for build:** Sonnet (planned on Opus 4.8) · **Date:** 2026-06-09


> **How to use this plan.** This file is self-contained: an executor agent with a fresh context can build from it without the planning conversation. Every step has a **pass/fail acceptance gate** (compile, `go vet`, or a test with a hand-computed expected value). Do not improvise scope. If a step's acceptance gate cannot be met, **stop and report** — do not skip ahead or invent APIs.

---

## 0. Global conventions (apply to every phase)

- **Language:** Go 1.25, module `github.com/vault-router-keeper`.
- **Build gotcha:** a `go.work` at `/mnt/adiii_dev` excludes this module → **always prefix Go commands with `GOWORK=off`**.
- **Acceptance commands (must all pass after every phase):**
  ```
  GOWORK=off go build ./...
  GOWORK=off go vet ./...
  GOWORK=off gofmt -l .        # must print NOTHING (all files formatted)
  GOWORK=off go test ./...
  ```
- **Phase 1 is dependency-free** (stdlib only — `math`, `context`, `log/slog`, `fmt`). Do **not** add external deps in Phase 1. Phases 2–3 may add `go-ethereum` and an HTTP client.
- **Tests are table-driven**, with hand-computed expected values (the anti-hallucination oracle). Float comparisons use a tolerance (`±0.01` unless stated).
- **Commit per completed step** (one logical unit), message ending with the repo's co-author trailer if configured.
- **Do not modify** the existing contracts in §1 — only add to them.

---

## 1. Existing contracts (READ-ONLY — build against these, do not change)

From `pkg/types/types.go`:
```go
type StrategyID [32]byte
type Bps uint16
const BpsDenominator Bps = 10_000

type StrategyState struct {
    ID            StrategyID
    TargetBps     Bps      // current on-chain target
    CurrentAssets *big.Int
    CapBps        Bps      // owner-set on-chain cap
    Quarantined   bool
    APY           float64  // off-chain yield signal; decimal fraction (0.06 = 6%); 0 if unknown
}

type VaultState struct {
    TotalAssets          *big.Int
    IdleAssets           *big.Int
    IdleReserveBps       Bps
    MaxRebalanceDeltaBps Bps   // per-call churn bound; 0 = unset (rely on-chain)
    Paused               bool
    Strategies           []StrategyState
    PendingWithdraws     []WithdrawRequest
}

type Allocation struct { Targets map[StrategyID]Bps }
```

From `internal/brain/decider.go` (the interface the brain must satisfy — unchanged):
```go
type Decider interface {
    Decide(ctx context.Context, state *types.VaultState) (*types.Allocation, error)
}
```
`StubDecider` already implements it (echoes current targets). The new brain is an **additional** implementation; `StubDecider` stays.

---

## 2. Orchestration model (how agents/models are assigned)

| Phase | What | Parallelism | Agent/model | Isolation |
|---|---|---|---|---|
| **1 — Brain core** | closed-form risk + risk-gated allocator | **Sequential, single builder** (one cohesive module — parallel builders would only collide) | 1× Sonnet builder | none (single branch) |
| **2 — Facet adapters** | Credora / Aave / Pendle readers | **Fan-out: 3 builders, one per facet** (independent files behind one interface) | 3× Sonnet builders | **worktree per builder** (parallel file edits) |
| **3 — Validation + live wiring** | safety stage + ChainReader + LocalExecutor | Sequential | 1× Sonnet builder | none |

**Planner (Opus) re-grounds between phases:** after each phase's tests are green, the next phase's plan is re-checked against what actually compiled (real API shapes, real type names). Do not plan Phase 2 details on assumptions Phase 1 disproved.

---

## 3. Phase 1 — Brain core (dependency-free, fully testable NOW)

Implements: the closed-form `EL = PD × LGD`, the `max(vendor, closed-form)` composition, and the risk-gated water-fill `Decider`.

> **Role of APY (yield is the objective; risk is the gate).** APY *is* the primary driver — the allocator's whole job is to maximize yield. Risk does **not** replace APY; the objective the water-fill maximizes is `effectiveYield(APY) - Lambda*risk_m`, i.e. **risk-adjusted** APY. Risk enters two ways: (a) it *penalizes* the score, and (b) it *caps* via `effectiveCapBps`. Two cautions baked in: **never chase raw spot APY** — it spikes, it's gameable, and a high APY is itself often a risk signal (the market pays more because it's riskier); so `effectiveYield` clamps `MaxPlausibleAPY` (a stand-in for Gauntlet's 30-day rolling smoothing). And rebalancing is **multi-trigger**: APY drift is the main one, but a risk spike or the Pendle liquidity gate can also force a move. Phase 1 uses *effective spot* APY with greedy fill-to-cap; true marginal-rate water-filling (rate falls as you supply more) is a documented Phase-2 refinement (see §7).

### 3.1 Files to create

**`internal/brain/risk.go`** — the risk core. Define:

```go
type RiskInputs struct {
    Sigma            float64 // annualized vol of collateral
    HorizonYears     float64 // e.g. 7.0/365
    LiqPriceRatio    float64 // S_liq/S0 in (0,1)
    Pegged           bool
    DepegWatch       bool
    LiquidityHaircut float64 // [0,1]; 1 = illiquid
    HasVendor        bool    // a facet adapter supplied a vendor EL
    VendorEL         float64 // vendor expected-loss in [0,1] (e.g. Credora PSL), composed via max()
    HasLiquidityCap  bool      // adapter set a hard liquidity ceiling (the Pendle queue gate, LLD PENGATE node)
    LiquidityCapBps  types.Bps // max target this strategy may hold given exit liquidity / maturity
}

type RiskProvider interface { Risk(id types.StrategyID) (RiskInputs, bool) } // ok=false => no data

type StaticRiskProvider struct{ /* map[types.StrategyID]RiskInputs */ }
func NewStaticRiskProvider(m map[types.StrategyID]RiskInputs) *StaticRiskProvider

type RiskModel struct {
    BroadShockVolatile, BroadShockPegged, DepegShock float64
    LiqPenaltyFloor, DepegPenaltyFloor               float64
    Lambda, SoftThreshold, KillThreshold             float64
    MaxPlausibleAPY, UnknownEL                        float64
    StepBps, MinDeltaBps                              types.Bps
    MinScore                                          float64
}
func DefaultRiskModel() RiskModel
```

Methods/functions (signatures fixed; bodies per §3.3 semantics):
```go
func (m RiskModel) effectiveYield(apy float64) float64
func (m RiskModel) expectedLoss(r RiskInputs) float64        // = max(closedFormEL, VendorEL), clamped [0,1]
func (m RiskModel) closedFormEL(r RiskInputs) float64        // = PD * max(LGD_broad, LGD_depeg)
func (m RiskModel) effectiveCapBps(ownerCap types.Bps, el float64) types.Bps
func pLiq(liqRatio, sigma, t float64) float64                // first-passage barrier prob
func severity(shock, buffer, haircut, floor float64) float64 // loss-given-liquidation
func normCDF(x float64) float64                              // 0.5*math.Erfc(-x/math.Sqrt2)
func clamp01(x float64) float64
```

`DefaultRiskModel()` values: BroadShockVolatile 0.40, BroadShockPegged 0.10, DepegShock 0.30, LiqPenaltyFloor 0.05, DepegPenaltyFloor 0.25, Lambda 1.0, SoftThreshold 0.02, KillThreshold 0.10, MaxPlausibleAPY 0.50, UnknownEL 0.0, StepBps 25, MinDeltaBps 50, MinScore math.Inf(-1).

**`internal/brain/gauntlet_lite.go`** — the allocator:
```go
type GauntletLiteDecider struct{ /* RiskProvider, RiskModel, *slog.Logger */ }
func NewGauntletLiteDecider(r RiskProvider, m RiskModel, log *slog.Logger) *GauntletLiteDecider
func (d *GauntletLiteDecider) Decide(ctx context.Context, state *types.VaultState) (*types.Allocation, error)
```

**`internal/brain/risk_test.go`** and **`internal/brain/gauntlet_lite_test.go`** — tests in §3.4.

### 3.2 Wiring (add a brain selector)

- `internal/config/config.go`: add field `Brain string` to `Config`; in `Load()` add `Brain: env("KEEPER_BRAIN", "stub")`.
- `cmd/keeper/main.go`: replace `decider := brain.NewStubDecider()` with a switch:
  ```go
  var decider brain.Decider
  switch cfg.Brain {
  case "gauntlet-lite":
      decider = brain.NewGauntletLiteDecider(brain.NewStaticRiskProvider(nil), brain.DefaultRiskModel(), log)
  default:
      decider = brain.NewStubDecider()
  }
  ```
- `.env.example`: add `KEEPER_BRAIN=stub` with a comment.

### 3.3 Semantics (exact behavior)

- **`normCDF(x)`** = `0.5 * math.Erfc(-x / math.Sqrt2)`.
- **`pLiq`**: if `liqRatio >= 1` → 1; if `liqRatio<=0 || sigma<=0 || t<=0` → 0; else `clamp01(2 * normCDF(ln(liqRatio)/(sigma*sqrt(t))))`.
- **`severity`**: if `haircut<=0` → 0; clamp haircut to [0,1]; `denom = max(1-buffer, 1e-9)`; `excess = clamp01((shock-buffer)/denom)`; `sev = floor + (1-floor)*excess`; return `clamp01(haircut*sev)`.
- **`closedFormEL`**: `pd = pLiq(...)`; `buffer = 1-LiqPriceRatio`; broad shock = Pegged ? BroadShockPegged : BroadShockVolatile; `lgd = severity(broad, buffer, haircut, LiqPenaltyFloor)`; if DepegWatch, `lgd = max(lgd, severity(DepegShock, buffer, haircut, DepegPenaltyFloor))`; return `clamp01(pd*lgd)`.
- **`expectedLoss`**: `el = closedFormEL(r)`; if `HasVendor && VendorEL>el` → `el=VendorEL`; return `clamp01(el)`.
- **`effectiveCapBps`**: `el>=Kill` → 0; `el<=Soft` → ownerCap; else `round(ownerCap * (1-(el-Soft)/(Kill-Soft)))`.
- **`effectiveYield`**: `apy<0`→0; clamp to MaxPlausibleAPY.
- **`Decide`** algorithm:
  1. For each strategy: if `Quarantined` → EL=1 (cap 0). Else `(ri,known)=provider.Risk(ID)`; `el = known ? expectedLoss(ri) : UnknownEL`. `cap_i = effectiveCapBps(CapBps, el)` (0 ⇒ killed). **Liquidity/Pendle gate (LLD PENGATE):** if `known && ri.HasLiquidityCap` → `cap_i = min(cap_i, ri.LiquidityCapBps)` — route an illiquid venue (e.g. Pendle PT) only up to what the async withdraw queue can cover before maturity. `yield_i = effectiveYield(APY)`.
  2. **Water-fill** the investable budget `= BpsDenominator - IdleReserveBps` in `StepBps` chunks: each chunk goes to the eligible strategy (alloc < `cap_i`, the risk- and liquidity-adjusted cap from step 1) with the highest marginal score `yield_i - Lambda*el_i`; stop when budget exhausted, no eligible capacity, or best score `< MinScore`. *(Concentration / group caps — LLD `WF` node — are deferred to Phase 2; see §7. `CapBps` bounds single-venue concentration in Phase 1.)*
  3. **Hysteresis:** if `|alloc_i - TargetBps_i| < MinDeltaBps` → set `alloc_i = TargetBps_i`.
  4. **Churn clamp:** if `MaxRebalanceDeltaBps > 0` and `Σ|alloc_i - TargetBps_i| > MaxRebalanceDeltaBps`, scale every delta by `MaxRebalanceDeltaBps / Σ|Δ|` toward current.
  5. Return `Allocation{Targets}` containing **every** strategy (killed ones = 0).

### 3.4 Acceptance gate — test vectors (hand-computed; these MUST pass)

`risk_test.go`:
- `normCDF(0)==0.5`; `normCDF(-1)≈0.1587`; `normCDF(1)≈0.8413` (±0.001).
- `pLiq(0.9, 1.0, 1.0/52) ≈ 0.447` (±0.01).
- `pLiq(0.5, 0.3, 1.0/52) ≈ 0` (±0.001); `pLiq(1.0, …) == 1`.
- `closedFormEL(RiskInputs{Sigma:1.0, HorizonYears:1.0/52, LiqPriceRatio:0.9, LiquidityHaircut:0.8})` with DefaultRiskModel ≈ **0.131** (assert 0.12–0.14).
- `effectiveCapBps(5000, 0.01)==5000`; `effectiveCapBps(5000, 0.06)==2500`; `effectiveCapBps(5000, 0.10)==0`.
- `expectedLoss(RiskInputs{Sigma:0, HasVendor:true, VendorEL:0.2})==0.2` (vendor dominates when closed-form is 0).

`gauntlet_lite_test.go` (use `StaticRiskProvider`; build `StrategyID`s via a helper that sets the first byte):
- **Ranking+caps+idle:** 3 strategies, all `CapBps:5000`, EL 0 (Sigma 0, no vendor), APY A=0.08 B=0.05 C=0.03, `IdleReserveBps:1000`, `MaxRebalanceDeltaBps:0`, targets 0 → expect **A=5000, B=4000, C=0**, sum=9000.
- **Kill switch:** strategy with `HasVendor:true, VendorEL:0.2` → its target == **0**.
- **Soft cap:** strategy with VendorEL 0.06, CapBps 5000 → effective cap 2500, so its alloc ≤ 2500.
- **Churn clamp:** desired A=5000 B=4000, `MaxRebalanceDeltaBps:900`, targets 0 → scaled to A=500 B=400 (Σ|Δ| ≤ 900).
- **Liquidity/Pendle gate:** strategy with `HasLiquidityCap:true, LiquidityCapBps:1500`, `CapBps:5000`, EL 0, highest APY → its target ≤ **1500** (the queue gate beats both the owner cap and the yield ranking).
- **Empty/edge:** `len(Strategies)==0` → empty `Targets`, no error; `Paused` is irrelevant to the brain (the keeper loop handles pause).

### 3.5 Phase 1 Definition of Done
All §0 acceptance commands pass; all §3.4 vectors green; `KEEPER_BRAIN=gauntlet-lite GOWORK=off go run ./cmd/keeper` starts and runs the loop (StubReader → empty state → no-op, clean SIGTERM). `StubDecider` still default.

---

## 4. Phase 2 — Facet adapters (fan-out, worktree-isolated)

Three independent `RiskProvider`/adapter implementations behind a router. **Network/chain calls live behind interfaces; tests use recorded fixtures, never live endpoints.** Each builder owns **one subpackage**; run in separate worktrees.

Finalized layout (ports & adapters):
```
internal/risk/                # ADAPTERS for brain.RiskProvider (the EL feeds)
  composite.go                #   package risk: routes StrategyID -> facet adapter
  credora/  aave/  pendle/    #   one subpackage per facet (independent, worktree-friendly)
internal/bindings/<contract>/ # abigen-generated Go bindings (vault, aavepool, chaosrisk, redstone)
internal/chain/               # ethclient wrapper + multicall (shared infra)
abi/<contract>.abi.json       # raw ABIs, source for `make bindings`
```

### 4.1 Files
- `internal/risk/composite.go` — package `risk`; `CompositeProvider` implementing `brain.RiskProvider`: routes each `StrategyID` to its facet adapter; returns `(_, false)` (⇒ UnknownEL → closed-form floor) when no adapter matches. Config = `map[StrategyID]facetKind`.
- `internal/risk/credora/` — **CredoraReader (Morpho)**: GraphQL client → PSL (D→A) + Stress/WithdrawLiquidity/Diversification → `RiskInputs{HasVendor:true, VendorEL:<PSL→EL>}`. PSL→EL mapping documented in code (A→~0, D→~0.15+). HTTP behind an interface; fixture-based test.
- `internal/risk/aave/` — **AaveChaosReader (Aave V3)**: on-chain reads (UiPoolDataProvider via `internal/bindings/aavepool`): utilization, available liquidity, paused/frozen, Chaos-set caps → `RiskInputs` (no vendor EL; util/paused floor + closed-form). Chain client behind an interface; mocked test.
- `internal/risk/pendle/` — **PendlePTReader (Pendle PT)**: PT oracle reads (Chaos killswitch getter on-chain; RedStone push adapter where available — *RedStone pull has no Go SDK, see §7*) → mark, time-to-maturity, killswitch → `VendorEL`. **Computes `LiquidityCapBps` (the Pendle queue gate)** from `VaultState.PendingWithdraws` + time-to-maturity: the PT target may not exceed what the async withdraw queue can cover before maturity. Mocked test, incl. a case where pending withdraws force `LiquidityCapBps` below `CapBps`.
- `internal/bindings/<contract>/` + `abi/<contract>.abi.json` — abigen bindings + raw ABIs (see `abi/README.md`).

### 4.2 Dependencies
- `GOWORK=off go get github.com/ethereum/go-ethereum` (for aave/pendle on-chain reads).
- abigen bindings into `internal/bindings/<contract>/` (raw ABIs in `abi/`, regenerate via `make bindings`) — only the views the readers need.
- HTTP/GraphQL: stdlib `net/http` + `encoding/json` (no heavy client needed).

### 4.3 Acceptance gate
- Each adapter has a unit test driven by a **recorded fixture** (JSON for Credora; mocked contract returns for Aave/Pendle) asserting the produced `RiskInputs`/`VendorEL`.
- `CompositeProvider` routing test: correct adapter per `StrategyID`; unknown ID → `(_,false)`.
- Composition end-to-end: a high-PSL Credora fixture drives a Morpho strategy's target toward 0 through `GauntletLiteDecider`.
- All §0 commands pass with the new deps.

### 4.4 Phase 2 DoD
Three adapters + composite build and test green; `CompositeProvider` wired as an option in `main.go` (e.g. `KEEPER_BRAIN=gauntlet-lite` + a provider selector); live calls still gated by `KEEPER_DRY_RUN`.

---

## 5. Phase 3 — Validation layer + live wiring (sequential, later)

- `internal/validate/` — safety stage modeled on the Chaos-Agents `check()→execute()` pattern (MIT, https://github.com/ChaosLabsInc/chaos-agents): given the brain's `Allocation` + `VaultState`, **re-check every bound off-chain** (caps, idle floor, Σ|Δ| ≤ churn), optional `eth_call`/Tenderly simulation, expiry + min-delay; on fail → skip + alert, fall back to last-good/idle. Sits between `brain.Decider` and `execute.Executor` in the keeper loop.
- `internal/perceive` — real `ChainReader` replacing `StubReader` (reads Allocator/Guard/WithdrawQueue views via bindings).
- `internal/execute` — `LocalExecutor` replacing `LogExecutor` when `KEEPER_DRY_RUN=false` (signs with curator key from `os.Getenv(cfg.KeyEnv)`); selectors limited to the curator set.
- **Acceptance:** validation rejects an out-of-bounds allocation in a unit test; dry-run path byte-for-byte unchanged; live path behind `KEEPER_DRY_RUN=false` only.

---

## 6. Invariants (must hold in every phase — these are acceptance criteria)

1. **Tightening-only** — a vendor signal can only *lower* an effective cap, never raise it past `CapBps`. (`expectedLoss` uses `max`; `effectiveCapBps` never exceeds `ownerCap`.)
2. **No single feed moves funds up** — composition floors at the closed-form; killed = 0; the brain never invents capacity.
3. **Structural caps are hard caps** — water-fill never exceeds `CapBps`, `LiquidityCapBps` (the Pendle queue gate), `BpsDenominator - IdleReserveBps`, or `MaxRebalanceDeltaBps`.
4. **Determinism** — same `(VaultState, RiskInputs, RiskModel)` → identical `Allocation` (no `Date.now`/random).
5. **The keeper loop stays dull** — all intelligence is in `brain`/`risk`; `internal/keeper` planning logic is untouched.

---

## 7. Out of scope / open questions (do NOT build; flag if hit)
- **Concentration / group caps** (e.g. non-blue-chip ≤ 40%, the LLD `WF` node) — **DECISION: deferred to Phase 2.** They need an asset-class tag absent from `StrategyState`; implementing now would require mocked/hallucinated classification (avoided per the grounding rule). Phase 1's per-strategy `CapBps` already bounds single-venue concentration. Phase 2 adds an asset-class input + a group-cap pass.
- **Marginal-rate water-filling** (rate decays as you supply more, true Idle-style water-fill) — Phase 2 refinement; Phase 1 uses effective-spot APY with greedy fill-to-cap.
- Exact PSL→EL and utilization→EL calibration curves (Phase 2 will need real fixtures to tune; use documented placeholders).
- Delegated-key/session-key execution stack (Phase 3+).
- Confirm the specific PT market the Pendle strategy holds has a live RedStone/Chaos oracle before relying on it.
- **RedStone *pull*-model feeds have no Go SDK** (the `@redstone-finance/evm-connector` is JS/TS only). Prefer the Chaos on-chain killswitch getter and/or a RedStone *push* adapter (Chainlink-compatible getter, Go-clean); use a thin TS sidecar only if a pull feed is unavoidable. Confirm the PT market's feed model before building `internal/risk/pendle`.
- Anything requiring a change to the §1 contracts → stop and report.
