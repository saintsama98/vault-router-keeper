package aave

import (
	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// ConfiguredReader supplies operator-set risk profiles for Aave V3 supply
// strategies.
//
// Grounding note (PLAN.md §4.1, re-scoped during Phase 2). The Aave facet in
// this stack reads its position via aToken.balanceOf only — IAavePool exposes
// just supply/withdraw, and there is NO UiPoolDataProvider / reserve-data getter
// in the diamond or its dependencies. A live utilization/paused/cap feed
// therefore needs an externally-sourced Aave V3 data-provider ABI (an open
// item). Until that lands, this adapter serves a conservative, operator-set
// RiskInputs per strategy. The brain composes it via max() with its closed-form
// EL, so a configured profile can only ever tighten an allocation, never loosen
// it (invariant 1).
type ConfiguredReader struct {
	profiles map[types.StrategyID]brain.RiskInputs
}

// NewConfiguredReader builds an Aave reader from per-strategy risk profiles.
func NewConfiguredReader(profiles map[types.StrategyID]brain.RiskInputs) *ConfiguredReader {
	return &ConfiguredReader{profiles: profiles}
}

func (r *ConfiguredReader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if r == nil || r.profiles == nil {
		return brain.RiskInputs{}, false
	}
	ri, ok := r.profiles[id]
	return ri, ok
}

// DefaultProfile is a conservative profile for an Aave V3 stablecoin supply
// position: pegged collateral with a small liquidity haircut standing in for the
// withdrawal/utilization risk that cannot currently be read live. Its closed-form
// EL is well below the soft threshold, so it does not gate a healthy market —
// it only documents the facet's risk posture and gives operators a tuning point.
func DefaultProfile() brain.RiskInputs {
	return brain.RiskInputs{
		Pegged:           true,
		LiqPriceRatio:    0.98,
		Sigma:            0.05,
		HorizonYears:     7.0 / 365,
		LiquidityHaircut: 0.10,
	}
}
