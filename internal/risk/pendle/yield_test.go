package pendle

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/vault-router-keeper/pkg/types"
)

func newYield(c Chain, id types.StrategyID) *YieldReader {
	return NewYieldReader(c, []types.StrategyID{id}, func() int64 { return fixedNow })
}

func TestYieldImpliedAPY(t *testing.T) {
	id := pid(1)
	// discount = 0.95, ttm = half a year => APY = (1/0.95)^2 - 1 ≈ 0.1080.
	half := int64(secondsPerYear / 2)
	c := mockChain{ready: true, rate: big.NewInt(95e16), expiry: fixedNow + half}
	apy, ok := newYield(c, id).APY(id)
	if !ok {
		t.Fatalf("APY ok=false, want true")
	}
	want := math.Pow(1/0.95, 2) - 1
	if math.Abs(apy-want) > 1e-9 {
		t.Errorf("APY = %v, want ~%v", apy, want)
	}
}

func TestYieldNotOwned(t *testing.T) {
	c := mockChain{ready: true, rate: big.NewInt(95e16), expiry: fixedNow + secondsPerYear}
	if _, ok := newYield(c, pid(1)).APY(pid(2)); ok {
		t.Errorf("unowned strategy ok=true, want false")
	}
}

func TestYieldDegradesToNoData(t *testing.T) {
	id := pid(1)
	cases := map[string]mockChain{
		"oracle not ready":      {ready: false, rate: big.NewInt(95e16), expiry: fixedNow + secondsPerYear},
		"oracle err":            {ready: true, readyErr: errors.New("x"), rate: big.NewInt(95e16), expiry: fixedNow + 1},
		"rate err":              {ready: true, rateErr: errors.New("x"), expiry: fixedNow + 1},
		"non-positive rate":     {ready: true, rate: big.NewInt(0), expiry: fixedNow + 1},
		"no discount (rate>=1)": {ready: true, rate: big.NewInt(1e18), expiry: fixedNow + secondsPerYear},
		"matured":               {ready: true, rate: big.NewInt(95e16), expiry: fixedNow - 1},
		"expiry err":            {ready: true, rate: big.NewInt(95e16), expiryErr: errors.New("x"), expiry: fixedNow + 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, ok := newYield(c, id).APY(id); ok {
				t.Errorf("%s: ok=true, want false (no overlay)", name)
			}
		})
	}
}

func TestYieldNilChain(t *testing.T) {
	yr := NewYieldReader(nil, []types.StrategyID{pid(1)}, func() int64 { return fixedNow })
	if _, ok := yr.APY(pid(1)); ok {
		t.Errorf("nil chain ok=true, want false")
	}
}
