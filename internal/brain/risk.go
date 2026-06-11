package brain

import (
	"math"

	"github.com/vault-router-keeper/pkg/types"
)

// RiskInputs is the per-strategy risk view a RiskProvider supplies. It is the
// off-chain signal layer the closed-form model consumes; a facet adapter
// (Phase 2) fills it from Credora / Aave / Pendle, while StaticRiskProvider
// serves fixed values for tests and the dependency-free Phase 1 build.
type RiskInputs struct {
	Sigma            float64 // annualized vol of collateral
	HorizonYears     float64 // e.g. 7.0/365
	LiqPriceRatio    float64 // S_liq/S0 in (0,1)
	Pegged           bool
	DepegWatch       bool
	LiquidityHaircut float64 // [0,1]; 1 = illiquid

	HasVendor bool    // a facet adapter supplied a vendor EL
	VendorEL  float64 // vendor expected-loss in [0,1] (e.g. Credora PSL), composed via max()

	HasLiquidityCap bool      // adapter set a hard liquidity ceiling (the Pendle queue gate)
	LiquidityCapBps types.Bps // max target this strategy may hold given exit liquidity / maturity
}

// RiskProvider maps a strategy id to its risk inputs. ok=false means no data,
// which the allocator treats as the closed-form floor (UnknownEL).
type RiskProvider interface {
	Risk(id types.StrategyID) (RiskInputs, bool)
}

// StaticRiskProvider serves risk inputs from an in-memory map. A nil map (or nil
// receiver) reports "no data" for every id.
type StaticRiskProvider struct {
	m map[types.StrategyID]RiskInputs
}

func NewStaticRiskProvider(m map[types.StrategyID]RiskInputs) *StaticRiskProvider {
	return &StaticRiskProvider{m: m}
}

func (p *StaticRiskProvider) Risk(id types.StrategyID) (RiskInputs, bool) {
	if p == nil || p.m == nil {
		return RiskInputs{}, false
	}
	r, ok := p.m[id]
	return r, ok
}

// RiskModel holds the closed-form parameters and the allocator's gating
// thresholds. Construct it with DefaultRiskModel and override fields as needed.
type RiskModel struct {
	BroadShockVolatile, BroadShockPegged, DepegShock float64
	LiqPenaltyFloor, DepegPenaltyFloor               float64
	Lambda, SoftThreshold, KillThreshold             float64
	MaxPlausibleAPY, UnknownEL                       float64
	StepBps, MinDeltaBps                             types.Bps
	MinScore                                         float64
}

// DefaultRiskModel returns the calibrated Phase 1 defaults (see PLAN.md §3.1).
func DefaultRiskModel() RiskModel {
	return RiskModel{
		BroadShockVolatile: 0.40,
		BroadShockPegged:   0.10,
		DepegShock:         0.30,
		LiqPenaltyFloor:    0.05,
		DepegPenaltyFloor:  0.25,
		Lambda:             1.0,
		SoftThreshold:      0.02,
		KillThreshold:      0.10,
		MaxPlausibleAPY:    0.50,
		UnknownEL:          0.0,
		StepBps:            25,
		MinDeltaBps:        50,
		MinScore:           math.Inf(-1),
	}
}

// normCDF is the standard-normal CDF, Φ(x) = 0.5·erfc(-x/√2).
func normCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}

// clamp01 bounds x to [0,1].
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// pLiq is the first-passage probability that the collateral price crosses the
// liquidation barrier S_liq/S0 = liqRatio within horizon t, under GBM with
// volatility sigma. liqRatio >= 1 means already at/over the barrier (PD 1);
// non-positive sigma/t or liqRatio means PD 0.
func pLiq(liqRatio, sigma, t float64) float64 {
	if liqRatio >= 1 {
		return 1
	}
	if liqRatio <= 0 || sigma <= 0 || t <= 0 {
		return 0
	}
	return clamp01(2 * normCDF(math.Log(liqRatio)/(sigma*math.Sqrt(t))))
}

// severity is the loss-given-liquidation: how much of the haircut is realized
// once a shock of size `shock` hits a position with collateral buffer `buffer`.
// A zero haircut means no loss; otherwise loss scales from `floor` up to the
// full haircut as the shock exceeds the buffer.
func severity(shock, buffer, haircut, floor float64) float64 {
	if haircut <= 0 {
		return 0
	}
	haircut = clamp01(haircut)
	denom := 1 - buffer
	if denom < 1e-9 {
		denom = 1e-9
	}
	excess := clamp01((shock - buffer) / denom)
	sev := floor + (1-floor)*excess
	return clamp01(haircut * sev)
}

// closedFormEL is the model's own expected loss, EL = PD × LGD, taking the worse
// of the broad-shock and (when watched) depeg-shock severities.
func (m RiskModel) closedFormEL(r RiskInputs) float64 {
	pd := pLiq(r.LiqPriceRatio, r.Sigma, r.HorizonYears)
	buffer := 1 - r.LiqPriceRatio
	broad := m.BroadShockVolatile
	if r.Pegged {
		broad = m.BroadShockPegged
	}
	lgd := severity(broad, buffer, r.LiquidityHaircut, m.LiqPenaltyFloor)
	if r.DepegWatch {
		if d := severity(m.DepegShock, buffer, r.LiquidityHaircut, m.DepegPenaltyFloor); d > lgd {
			lgd = d
		}
	}
	return clamp01(pd * lgd)
}

// expectedLoss composes the closed-form EL with any vendor EL via max() — a
// vendor feed can only raise loss (tighten), never lower it.
func (m RiskModel) expectedLoss(r RiskInputs) float64 {
	el := m.closedFormEL(r)
	if r.HasVendor && r.VendorEL > el {
		el = r.VendorEL
	}
	return clamp01(el)
}

// effectiveCapBps maps an owner cap and an expected loss to the risk-adjusted
// cap: full cap below the soft threshold, zero at/above the kill threshold, and
// a linear taper in between. Output never exceeds ownerCap.
func (m RiskModel) effectiveCapBps(ownerCap types.Bps, el float64) types.Bps {
	if el >= m.KillThreshold {
		return 0
	}
	if el <= m.SoftThreshold {
		return ownerCap
	}
	frac := 1 - (el-m.SoftThreshold)/(m.KillThreshold-m.SoftThreshold)
	return types.Bps(math.Round(float64(ownerCap) * frac))
}

// effectiveYield clamps a raw APY: negatives become 0, and the value is capped
// at MaxPlausibleAPY so a gamed/spiking spot rate cannot seduce the allocator.
func (m RiskModel) effectiveYield(apy float64) float64 {
	if apy < 0 {
		return 0
	}
	if apy > m.MaxPlausibleAPY {
		return m.MaxPlausibleAPY
	}
	return apy
}
