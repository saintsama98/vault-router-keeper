package brain

import (
	"context"
	"testing"

	"github.com/vault-router-keeper/pkg/types"
)

// sid builds a StrategyID with a distinguishing first byte.
func sid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

// newDecider wires a GauntletLiteDecider over the given static risk map.
func newDecider(m map[types.StrategyID]RiskInputs) *GauntletLiteDecider {
	return NewGauntletLiteDecider(NewStaticRiskProvider(m), DefaultRiskModel(), nil)
}

func sumTargets(a *types.Allocation) types.Bps {
	var s types.Bps
	for _, v := range a.Targets {
		s += v
	}
	return s
}

func TestDecideRankingCapsIdle(t *testing.T) {
	d := newDecider(nil) // no risk data → EL 0 for all
	state := &types.VaultState{
		IdleReserveBps: 1000,
		Strategies: []types.StrategyState{
			{ID: sid(1), CapBps: 5000, APY: 0.08},
			{ID: sid(2), CapBps: 5000, APY: 0.05},
			{ID: sid(3), CapBps: 5000, APY: 0.03},
		},
	}
	a, err := d.Decide(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Targets[sid(1)]; got != 5000 {
		t.Errorf("A = %v, want 5000", got)
	}
	if got := a.Targets[sid(2)]; got != 4000 {
		t.Errorf("B = %v, want 4000", got)
	}
	if got := a.Targets[sid(3)]; got != 0 {
		t.Errorf("C = %v, want 0", got)
	}
	if got := sumTargets(a); got != 9000 {
		t.Errorf("sum = %v, want 9000 (10000 − 1000 idle)", got)
	}
}

func TestDecideKillSwitch(t *testing.T) {
	d := newDecider(map[types.StrategyID]RiskInputs{
		sid(9): {HasVendor: true, VendorEL: 0.2}, // ≥ kill threshold
	})
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: sid(9), CapBps: 5000, APY: 0.10},
	}}
	a, _ := d.Decide(context.Background(), state)
	if got := a.Targets[sid(9)]; got != 0 {
		t.Errorf("killed target = %v, want 0", got)
	}
}

func TestDecideSoftCap(t *testing.T) {
	d := newDecider(map[types.StrategyID]RiskInputs{
		sid(5): {HasVendor: true, VendorEL: 0.06}, // → effective cap 2500
	})
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: sid(5), CapBps: 5000, APY: 0.10},
	}}
	a, _ := d.Decide(context.Background(), state)
	if got := a.Targets[sid(5)]; got > 2500 {
		t.Errorf("soft-capped target = %v, want ≤ 2500", got)
	}
}

func TestDecideChurnClamp(t *testing.T) {
	d := newDecider(nil)
	state := &types.VaultState{
		IdleReserveBps:       1000,
		MaxRebalanceDeltaBps: 900,
		Strategies: []types.StrategyState{
			{ID: sid(1), CapBps: 5000, APY: 0.08},
			{ID: sid(2), CapBps: 5000, APY: 0.05},
		},
	}
	a, _ := d.Decide(context.Background(), state)
	// Desired A=5000 B=4000 (Σ|Δ|=9000) scaled by 900/9000 = 0.1.
	if got := a.Targets[sid(1)]; got != 500 {
		t.Errorf("A = %v, want 500", got)
	}
	if got := a.Targets[sid(2)]; got != 400 {
		t.Errorf("B = %v, want 400", got)
	}
}

func TestDecideLiquidityGate(t *testing.T) {
	d := newDecider(map[types.StrategyID]RiskInputs{
		sid(7): {HasLiquidityCap: true, LiquidityCapBps: 1500}, // queue gate
	})
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: sid(7), CapBps: 5000, APY: 0.20}, // highest APY, but gated
		{ID: sid(8), CapBps: 5000, APY: 0.01},
	}}
	a, _ := d.Decide(context.Background(), state)
	if got := a.Targets[sid(7)]; got > 1500 {
		t.Errorf("liquidity-gated target = %v, want ≤ 1500", got)
	}
}

func TestDecideEmpty(t *testing.T) {
	d := newDecider(nil)
	a, err := d.Decide(context.Background(), &types.VaultState{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Targets) != 0 {
		t.Errorf("targets = %d, want 0", len(a.Targets))
	}
}
