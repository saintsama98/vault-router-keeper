package aave

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/pkg/types"
)

// ray scales a base-10 exponent into a *big.Int RAY (1e27) value: rayPow(25)
// == 1e25, so 5*rayPow(25) == 5e25 ray == 0.05 fraction.
func rayPow(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}

// mockLiquidityRate is a canned liquidityRateCaller returning a fixed
// LiquidityRate (and optional error) for any asset. It never touches a chain.
type mockLiquidityRate struct {
	liquidityRate *big.Int
	err           error
}

func (m mockLiquidityRate) GetReserveData(_ *bind.CallOpts, _ common.Address) (struct {
	Unbacked                *big.Int
	AccruedToTreasuryScaled *big.Int
	TotalAToken             *big.Int
	TotalStableDebt         *big.Int
	TotalVariableDebt       *big.Int
	LiquidityRate           *big.Int
	VariableBorrowRate      *big.Int
	StableBorrowRate        *big.Int
	AverageStableBorrowRate *big.Int
	LiquidityIndex          *big.Int
	VariableBorrowIndex     *big.Int
	LastUpdateTimestamp     *big.Int
}, error) {
	out := struct {
		Unbacked                *big.Int
		AccruedToTreasuryScaled *big.Int
		TotalAToken             *big.Int
		TotalStableDebt         *big.Int
		TotalVariableDebt       *big.Int
		LiquidityRate           *big.Int
		VariableBorrowRate      *big.Int
		StableBorrowRate        *big.Int
		AverageStableBorrowRate *big.Int
		LiquidityIndex          *big.Int
		VariableBorrowIndex     *big.Int
		LastUpdateTimestamp     *big.Int
	}{}
	out.LiquidityRate = m.liquidityRate
	return out, m.err
}

func TestYieldReaderAPY(t *testing.T) {
	owned := types.StrategyID{0xaa}
	other := types.StrategyID{0xbb}
	asset := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	isAave := map[types.StrategyID]bool{owned: true}

	cases := []struct {
		name    string
		caller  liquidityRateCaller
		nilFrom bool // build via NewYieldReader with nil backend (unconfigured)
		id      types.StrategyID
		wantOK  bool
		wantAPY float64
	}{
		{
			// 5e25 ray / 1e27 == 0.05 == 5% supply APR.
			name:    "owned 5e25 ray -> 0.05",
			caller:  mockLiquidityRate{liquidityRate: new(big.Int).Mul(big.NewInt(5), rayPow(25))},
			id:      owned,
			wantOK:  true,
			wantAPY: 0.05,
		},
		{
			// 3.5e25 ray / 1e27 == 0.035 == 3.5% (35e24 == 3.5e25).
			name:    "owned 3.5e25 ray -> 0.035",
			caller:  mockLiquidityRate{liquidityRate: new(big.Int).Mul(big.NewInt(35), rayPow(24))},
			id:      owned,
			wantOK:  true,
			wantAPY: 0.035,
		},
		{
			name:   "not owned -> ok=false",
			caller: mockLiquidityRate{liquidityRate: new(big.Int).Mul(big.NewInt(5), rayPow(25))},
			id:     other,
			wantOK: false,
		},
		{
			name:   "zero rate -> ok=false",
			caller: mockLiquidityRate{liquidityRate: big.NewInt(0)},
			id:     owned,
			wantOK: false,
		},
		{
			name:   "nil rate -> ok=false",
			caller: mockLiquidityRate{liquidityRate: nil},
			id:     owned,
			wantOK: false,
		},
		{
			name:   "read error -> ok=false",
			caller: mockLiquidityRate{liquidityRate: new(big.Int).Mul(big.NewInt(5), rayPow(25)), err: errors.New("rpc down")},
			id:     owned,
			wantOK: false,
		},
		{
			name:    "unconfigured backend -> ok=false",
			nilFrom: true,
			id:      owned,
			wantOK:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r *YieldReader
			if tc.nilFrom {
				// nil backend => unconfigured no-data path.
				r = NewYieldReader(nil, common.Address{}, asset, isAave)
			} else {
				r = newYieldReaderFromCaller(tc.caller, asset, isAave)
			}
			apy, ok := r.APY(tc.id)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if math.Abs(apy-tc.wantAPY) > 1e-12 {
				t.Fatalf("apy = %v, want %v", apy, tc.wantAPY)
			}
		})
	}
}
