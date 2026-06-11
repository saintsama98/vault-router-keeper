package aave

import (
	"testing"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

func aid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

func TestConfiguredReader(t *testing.T) {
	want := brain.RiskInputs{Pegged: true, HasVendor: true, VendorEL: 0.04}
	r := NewConfiguredReader(map[types.StrategyID]brain.RiskInputs{aid(1): want})

	got, ok := r.Risk(aid(1))
	if !ok {
		t.Fatalf("configured id: ok=false, want true")
	}
	if got != want {
		t.Errorf("profile = %+v, want %+v", got, want)
	}

	if _, ok := r.Risk(aid(2)); ok {
		t.Errorf("unconfigured id: ok=true, want false")
	}
}

func TestNilReaderSafe(t *testing.T) {
	var r *ConfiguredReader
	if _, ok := r.Risk(aid(1)); ok {
		t.Errorf("nil reader: ok=true, want false")
	}
}

// DefaultProfile must sit below the soft threshold so a healthy Aave market is
// not gated by the conservative stand-in profile.
func TestDefaultProfileIsLowEL(t *testing.T) {
	m := brain.DefaultRiskModel()
	r := NewConfiguredReader(map[types.StrategyID]brain.RiskInputs{aid(1): DefaultProfile()})
	dec := brain.NewGauntletLiteDecider(r, m, nil)
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: aid(1), CapBps: 5000, APY: 0.05},
	}}
	a, err := dec.Decide(nil, state)
	if err != nil {
		t.Fatal(err)
	}
	// Low EL ⇒ full owner cap reachable; budget is 10000 with no idle reserve.
	if got := a.Targets[aid(1)]; got != 5000 {
		t.Errorf("default-profile target = %v, want 5000 (EL below soft threshold)", got)
	}
}
