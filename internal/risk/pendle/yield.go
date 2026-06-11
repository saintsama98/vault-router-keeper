package pendle

import (
	"math"
	"math/big"
	"time"

	"github.com/vault-router-keeper/pkg/types"
)

// rayLikeScale is the 1e18 fixed-point scale of the Pendle PT→asset rate.
var ptRateScale = new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

// weiRateToFloat converts a 1e18-scaled rate to a float64 discount factor.
func weiRateToFloat(rate *big.Int) float64 {
	d, _ := new(big.Float).Quo(new(big.Float).SetInt(rate), ptRateScale).Float64()
	return d
}

// secondsPerYear is the compounding base for the fixed-to-maturity implied APY
// (CALIBRATION TODO: 365d vs 365.25d / act-360 is a modeling choice). The INPUTS
// — PT→asset rate and PT expiry — are 100% live on-chain reads; only this
// constant and the annualization formula are documented defaults, not data.
const secondsPerYear = 365 * 24 * 60 * 60

// YieldReader implements perceive.YieldProvider for Pendle PT strategies. It
// derives the fixed-to-maturity implied APY from the live oracle: a PT trades
// below its underlying asset pre-maturity (discount = rate/1e18 < 1), and
// redeems 1:1 at expiry, so holding to maturity yields (1/discount)^(yr/ttm) - 1.
//
// It reuses the existing on-chain Chain surface (OracleReady / PtToAssetRate /
// Expiry) — no new ABI or binding. Any failure (not owned, oracle un-ready,
// non-positive/unreadable rate, non-positive time-to-maturity) returns ok=false
// so the OverlayReader leaves APY at 0 (the brain then water-fills by EL alone).
// This NEVER fabricates a yield: every number traces to a live read.
type YieldReader struct {
	chain Chain
	owns  map[types.StrategyID]struct{}
	now   func() int64
}

// NewYieldReader builds a Pendle implied-APY provider for the given strategy ids.
// now defaults to the wall clock; tests inject a fixed clock for determinism.
func NewYieldReader(chain Chain, ids []types.StrategyID, now func() int64) *YieldReader {
	owns := make(map[types.StrategyID]struct{}, len(ids))
	for _, id := range ids {
		owns[id] = struct{}{}
	}
	if now == nil {
		now = func() int64 { return time.Now().Unix() }
	}
	return &YieldReader{chain: chain, owns: owns, now: now}
}

// APY returns the live fixed-to-maturity implied APY for a Pendle PT strategy.
func (r *YieldReader) APY(id types.StrategyID) (float64, bool) {
	if r == nil || r.chain == nil {
		return 0, false
	}
	if _, ok := r.owns[id]; !ok {
		return 0, false
	}
	// Safety gate: only trust the venue if its TWAP oracle is seeded.
	if ready, err := r.chain.OracleReady(id); err != nil || !ready {
		return 0, false
	}
	rate, err := r.chain.PtToAssetRate(id)
	if err != nil || rate == nil || rate.Sign() <= 0 {
		return 0, false
	}
	exp, err := r.chain.Expiry(id)
	if err != nil || exp == nil || !exp.IsInt64() {
		return 0, false
	}
	ttm := exp.Int64() - r.now()
	if ttm <= 0 {
		// Expired/at-maturity PT redeems 1:1 — no forward fixed yield to overlay.
		return 0, false
	}

	// discount = rate / 1e18 (PT→asset rate is 1e18-scaled). discount in (0,1]
	// pre-maturity; >= 1 means no discount, so no positive implied yield.
	discount := weiRateToFloat(rate)
	if discount <= 0 || discount >= 1 {
		return 0, false
	}
	apy := math.Pow(1/discount, float64(secondsPerYear)/float64(ttm)) - 1
	if apy <= 0 || math.IsNaN(apy) || math.IsInf(apy, 0) {
		return 0, false
	}
	return apy, true
}
