package pendle

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

const fixedNow int64 = 1_000_000

func pid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

// mockChain returns canned reads for a single configured strategy.
type mockChain struct {
	ready     bool
	rate      *big.Int
	expiry    int64
	readyErr  error
	rateErr   error
	expiryErr error
}

func (m mockChain) OracleReady(types.StrategyID) (bool, error) { return m.ready, m.readyErr }
func (m mockChain) PtToAssetRate(types.StrategyID) (*big.Int, error) {
	return m.rate, m.rateErr
}
func (m mockChain) Expiry(types.StrategyID) (*big.Int, error) {
	return big.NewInt(m.expiry), m.expiryErr
}

func newReader(c Chain, id types.StrategyID) *Reader {
	return NewReader(c, DefaultPolicy(), []types.StrategyID{id}, func() int64 { return fixedNow })
}

func TestLongDatedPTTightCap(t *testing.T) {
	id := pid(1)
	// time-to-maturity == MaxHorizon ⇒ tightest cap == MinPTBps.
	c := mockChain{ready: true, rate: big.NewInt(95e16), expiry: fixedNow + DefaultPolicy().MaxHorizonSecs}
	ri, ok := newReader(c, id).Risk(id)
	if !ok {
		t.Fatal("ok=false, want true")
	}
	if !ri.HasLiquidityCap || ri.LiquidityCapBps != 500 {
		t.Errorf("LiquidityCap = (%v,%v), want (true,500)", ri.HasLiquidityCap, ri.LiquidityCapBps)
	}
}

func TestNearMaturityLooseCap(t *testing.T) {
	id := pid(1)
	c := mockChain{ready: true, rate: big.NewInt(99e16), expiry: fixedNow + 1} // ~0 horizon
	ri, ok := newReader(c, id).Risk(id)
	if !ok {
		t.Fatal("ok=false, want true")
	}
	if !ri.HasLiquidityCap || ri.LiquidityCapBps != 5000 {
		t.Errorf("LiquidityCap = (%v,%v), want (true,5000)", ri.HasLiquidityCap, ri.LiquidityCapBps)
	}
}

func TestOracleNotReadyDegrades(t *testing.T) {
	id := pid(1)
	c := mockChain{ready: false, rate: big.NewInt(95e16), expiry: fixedNow + 1000}
	if _, ok := newReader(c, id).Risk(id); ok {
		t.Errorf("not-ready oracle: ok=true, want false")
	}
}

func TestBadReadsDegrade(t *testing.T) {
	id := pid(1)
	cases := map[string]mockChain{
		"oracle err": {ready: true, readyErr: errors.New("x"), rate: big.NewInt(1), expiry: fixedNow + 1000},
		"rate err":   {ready: true, rate: big.NewInt(1), rateErr: errors.New("x"), expiry: fixedNow + 1000},
		"zero rate":  {ready: true, rate: big.NewInt(0), expiry: fixedNow + 1000},
		"expiry err": {ready: true, rate: big.NewInt(1), expiry: fixedNow + 1000, expiryErr: errors.New("x")},
	}
	for name, c := range cases {
		if _, ok := newReader(c, id).Risk(id); ok {
			t.Errorf("%s: ok=true, want false", name)
		}
	}
}

func TestExpiredPTNoGate(t *testing.T) {
	id := pid(1)
	c := mockChain{ready: true, rate: big.NewInt(1e18), expiry: fixedNow - 1} // already matured
	ri, ok := newReader(c, id).Risk(id)
	if !ok {
		t.Fatal("ok=false, want true")
	}
	if ri.HasLiquidityCap {
		t.Errorf("expired PT: HasLiquidityCap=true, want false (redeems 1:1)")
	}
}

func TestUnownedStrategy(t *testing.T) {
	c := mockChain{ready: true, rate: big.NewInt(1), expiry: fixedNow + 1000}
	if _, ok := newReader(c, pid(1)).Risk(pid(9)); ok {
		t.Errorf("unowned id: ok=true, want false")
	}
}

// End-to-end: a long-dated PT with the highest APY is still capped to the
// structural liquidity ceiling through the real GauntletLiteDecider.
func TestLiquidityGateThroughBrain(t *testing.T) {
	id := pid(1)
	c := mockChain{ready: true, rate: big.NewInt(95e16), expiry: fixedNow + DefaultPolicy().MaxHorizonSecs}
	dec := brain.NewGauntletLiteDecider(newReader(c, id), brain.DefaultRiskModel(), nil)
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: id, CapBps: 5000, APY: 0.30}, // most attractive, but illiquid
	}}
	a, err := dec.Decide(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Targets[id]; got > 500 {
		t.Errorf("long-dated PT target = %v, want ≤ 500 (queue gate < owner cap)", got)
	}
}
