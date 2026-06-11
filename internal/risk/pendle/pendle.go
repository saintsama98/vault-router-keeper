package pendle

import (
	"math"
	"math/big"
	"time"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// Chain is the narrow on-chain read surface the Pendle reader needs, satisfied
// in production by the abigen bindings (see onchain.go) and mocked in tests.
type Chain interface {
	OracleReady(id types.StrategyID) (bool, error)       // market TWAP oracle seeded enough to read?
	PtToAssetRate(id types.StrategyID) (*big.Int, error) // PT→asset TWAP, 1e18-scaled
	Expiry(id types.StrategyID) (*big.Int, error)        // PT maturity, unix seconds
}

// Policy is the structural liquidity-cap calibration for PT venues (placeholder
// values per PLAN.md §7 — tune against real exit-liquidity data).
type Policy struct {
	MaxHorizonSecs int64     // at/above this time-to-maturity the PT is treated as most illiquid
	MaxPTBps       types.Bps // loosest cap (near maturity)
	MinPTBps       types.Bps // tightest cap (far from maturity)
}

func DefaultPolicy() Policy {
	return Policy{
		MaxHorizonSecs: 180 * 24 * 60 * 60, // 180 days
		MaxPTBps:       5000,
		MinPTBps:       500,
	}
}

// liquidityCapBps tightens linearly from MaxPTBps (at maturity) to MinPTBps (at
// or beyond MaxHorizonSecs): the further a PT is from redeeming 1:1, the less of
// it the vault should hold, since it can only be exited early via the AMM.
func (p Policy) liquidityCapBps(ttmSecs int64) types.Bps {
	frac := float64(ttmSecs) / float64(p.MaxHorizonSecs)
	if frac > 1 {
		frac = 1
	}
	if frac < 0 {
		frac = 0
	}
	span := float64(p.MaxPTBps) - float64(p.MinPTBps)
	return types.Bps(math.Round(float64(p.MaxPTBps) - frac*span))
}

// Reader implements brain.RiskProvider for Pendle PT strategies. Its primary
// signal is the liquidity gate (LiquidityCapBps): a PT cannot be redeemed at par
// before maturity, so the further from maturity, the tighter the structural cap.
// The oracle-readiness check is a safety gate — an unseeded TWAP can revert, so
// an un-ready or unreadable market degrades to "no data" (the closed-form floor),
// never a cap raise.
//
// Scope note (PLAN.md §4.1): the richer pending-withdraw-aware cap — sizing the
// PT to what the async WithdrawQueue can cover before maturity — needs live queue
// demand, which the fixed brain.RiskProvider signature Risk(id) does not carry.
// Phase 2 ships the maturity-structural cap; wiring live queue pressure is an
// open interface decision (see the Phase 2 build report).
type Reader struct {
	chain  Chain
	policy Policy
	owns   map[types.StrategyID]struct{}
	now    func() int64
}

// NewReader builds a Pendle reader for the given strategy ids. now defaults to
// the wall clock; tests inject a fixed clock for determinism.
func NewReader(chain Chain, policy Policy, ids []types.StrategyID, now func() int64) *Reader {
	owns := make(map[types.StrategyID]struct{}, len(ids))
	for _, id := range ids {
		owns[id] = struct{}{}
	}
	if now == nil {
		now = func() int64 { return time.Now().Unix() }
	}
	return &Reader{chain: chain, policy: policy, owns: owns, now: now}
}

func (r *Reader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if _, ok := r.owns[id]; !ok {
		return brain.RiskInputs{}, false
	}
	// Safety gate: only trust the venue if its TWAP oracle is seeded.
	if ready, err := r.chain.OracleReady(id); err != nil || !ready {
		return brain.RiskInputs{}, false
	}
	// Liveness/sanity: a non-positive or unreadable mark means we can't trust it.
	if rate, err := r.chain.PtToAssetRate(id); err != nil || rate == nil || rate.Sign() <= 0 {
		return brain.RiskInputs{}, false
	}
	exp, err := r.chain.Expiry(id)
	if err != nil || exp == nil || !exp.IsInt64() {
		return brain.RiskInputs{}, false
	}
	ttm := exp.Int64() - r.now()
	if ttm <= 0 {
		// Expired PT redeems 1:1 — fully liquid, no structural gate.
		return brain.RiskInputs{}, true
	}
	return brain.RiskInputs{
		HasLiquidityCap: true,
		LiquidityCapBps: r.policy.liquidityCapBps(ttm),
	}, true
}
