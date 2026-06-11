package aave

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// usdc scales a whole-token amount into 6-decimal base units (Aave USDC reserve).
func usdc(whole int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(whole), big.NewInt(1_000_000))
}

// mockReserve is a canned reserveCaller: it returns fixed reserve state for any
// asset, with optional per-call errors. It never touches a chain.
type mockReserve struct {
	totalAToken       *big.Int
	totalStableDebt   *big.Int
	totalVariableDebt *big.Int
	decimals          *big.Int
	supplyCap         *big.Int
	isActive          bool
	isFrozen          bool
	paused            bool

	dataErr, cfgErr, capsErr, pausedErr error
}

func (m mockReserve) GetReserveData(_ *bind.CallOpts, _ common.Address) (struct {
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
	}{
		TotalAToken:       m.totalAToken,
		TotalStableDebt:   m.totalStableDebt,
		TotalVariableDebt: m.totalVariableDebt,
	}
	return out, m.dataErr
}

func (m mockReserve) GetReserveConfigurationData(_ *bind.CallOpts, _ common.Address) (struct {
	Decimals                 *big.Int
	Ltv                      *big.Int
	LiquidationThreshold     *big.Int
	LiquidationBonus         *big.Int
	ReserveFactor            *big.Int
	UsageAsCollateralEnabled bool
	BorrowingEnabled         bool
	StableBorrowRateEnabled  bool
	IsActive                 bool
	IsFrozen                 bool
}, error) {
	out := struct {
		Decimals                 *big.Int
		Ltv                      *big.Int
		LiquidationThreshold     *big.Int
		LiquidationBonus         *big.Int
		ReserveFactor            *big.Int
		UsageAsCollateralEnabled bool
		BorrowingEnabled         bool
		StableBorrowRateEnabled  bool
		IsActive                 bool
		IsFrozen                 bool
	}{
		Decimals: m.decimals,
		IsActive: m.isActive,
		IsFrozen: m.isFrozen,
	}
	return out, m.cfgErr
}

func (m mockReserve) GetReserveCaps(_ *bind.CallOpts, _ common.Address) (struct {
	BorrowCap *big.Int
	SupplyCap *big.Int
}, error) {
	return struct {
		BorrowCap *big.Int
		SupplyCap *big.Int
	}{SupplyCap: m.supplyCap}, m.capsErr
}

func (m mockReserve) GetPaused(_ *bind.CallOpts, _ common.Address) (bool, error) {
	return m.paused, m.pausedErr
}

func aaveID(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

var usdcAsset = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

func reader(m reserveCaller) *AaveChaosReader {
	return newAaveChaosReaderFromCaller(m, usdcAsset, map[types.StrategyID]bool{aaveID(1): true})
}

const floatTol = 1e-9

func TestAaveChaosReader(t *testing.T) {
	// decimals=6, supplyCap=2,000,000 whole USDC => capBase=2e6*1e6.
	cases := []struct {
		name string
		m    mockReserve
		want brain.RiskInputs
	}{
		{
			// util = 500k/1M = 0.50 -> no haircut. headroom: supplied 1M of 2M cap
			// => 5000 bps. vendorEL = base (util < u0). Active, not frozen/paused.
			name: "healthy low util",
			m: mockReserve{
				totalAToken:       usdc(1_000_000),
				totalVariableDebt: usdc(500_000),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(2_000_000),
				isActive:          true,
			},
			want: brain.RiskInputs{
				HasLiquidityCap: true,
				LiquidityCapBps: 5000,
				HasVendor:       true,
				VendorEL:        0.0005,
			},
		},
		{
			// util = 950k/1M = 0.95 -> mid band: frac=(0.95-0.90)/0.09=0.5556,
			// haircut=0.5556*0.50=0.277778. supplyCap=0 -> no cap. vendorEL =
			// 0.0005 + 0.05*(0.95-0.80) = 0.008.
			name: "high util raises haircut",
			m: mockReserve{
				totalAToken:       usdc(1_000_000),
				totalVariableDebt: usdc(950_000),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(0),
				isActive:          true,
			},
			want: brain.RiskInputs{
				LiquidityHaircut: (0.05 / 0.09) * 0.50,
				HasVendor:        true,
				VendorEL:         0.008,
			},
		},
		{
			// util = 995k/1M = 0.995 > 0.99 -> haircut=0.80. vendorEL =
			// 0.0005 + 0.05*(0.995-0.80) = 0.0005 + 0.00975 = 0.01025.
			name: "near-100pct util top band",
			m: mockReserve{
				totalAToken:       usdc(1_000_000),
				totalVariableDebt: usdc(995_000),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(0),
				isActive:          true,
			},
			want: brain.RiskInputs{
				LiquidityHaircut: 0.80,
				HasVendor:        true,
				VendorEL:         0.01025,
			},
		},
		{
			// Frozen reserve: haircut floored at 0.50, DepegWatch set. util=0.5 so
			// utilHaircut=0 (max keeps 0.50). vendorEL base then floored at 0.02.
			name: "frozen winds down",
			m: mockReserve{
				totalAToken:       usdc(1_000_000),
				totalVariableDebt: usdc(500_000),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(0),
				isActive:          true,
				isFrozen:          true,
			},
			want: brain.RiskInputs{
				LiquidityHaircut: 0.50,
				DepegWatch:       true,
				HasVendor:        true,
				VendorEL:         0.02,
			},
		},
		{
			// Paused: haircut=1.0 (cannot exit). vendorEL floored at 0.02.
			name: "paused fully illiquid",
			m: mockReserve{
				totalAToken:       usdc(1_000_000),
				totalVariableDebt: usdc(500_000),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(0),
				isActive:          true,
				paused:            true,
			},
			want: brain.RiskInputs{
				LiquidityHaircut: 1.0,
				HasVendor:        true,
				VendorEL:         0.02,
			},
		},
		{
			// At supply cap: supplied == capBase -> 0 bps headroom (deposits revert).
			name: "supply cap exhausted",
			m: mockReserve{
				totalAToken:       usdc(2_000_000),
				totalVariableDebt: usdc(0),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(2_000_000),
				isActive:          true,
			},
			want: brain.RiskInputs{
				HasLiquidityCap: true,
				LiquidityCapBps: 0,
				HasVendor:       true,
				VendorEL:        0.0005,
			},
		},
		{
			// Empty reserve: totalAToken=0 -> util=0 (guarded), no haircut.
			name: "empty reserve zero util",
			m: mockReserve{
				totalAToken:       usdc(0),
				totalVariableDebt: usdc(0),
				decimals:          big.NewInt(6),
				supplyCap:         big.NewInt(0),
				isActive:          true,
			},
			want: brain.RiskInputs{
				HasVendor: true,
				VendorEL:  0.0005,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := reader(tc.m).Risk(aaveID(1))
			if !ok {
				t.Fatalf("ok=false, want true")
			}
			assertRiskInputs(t, got, tc.want)
		})
	}
}

func assertRiskInputs(t *testing.T, got, want brain.RiskInputs) {
	t.Helper()
	if math.Abs(got.LiquidityHaircut-want.LiquidityHaircut) > floatTol {
		t.Errorf("LiquidityHaircut = %v, want %v", got.LiquidityHaircut, want.LiquidityHaircut)
	}
	if got.DepegWatch != want.DepegWatch {
		t.Errorf("DepegWatch = %v, want %v", got.DepegWatch, want.DepegWatch)
	}
	if got.HasLiquidityCap != want.HasLiquidityCap {
		t.Errorf("HasLiquidityCap = %v, want %v", got.HasLiquidityCap, want.HasLiquidityCap)
	}
	if got.LiquidityCapBps != want.LiquidityCapBps {
		t.Errorf("LiquidityCapBps = %v, want %v", got.LiquidityCapBps, want.LiquidityCapBps)
	}
	if got.HasVendor != want.HasVendor {
		t.Errorf("HasVendor = %v, want %v", got.HasVendor, want.HasVendor)
	}
	if math.Abs(got.VendorEL-want.VendorEL) > floatTol {
		t.Errorf("VendorEL = %v, want %v", got.VendorEL, want.VendorEL)
	}
	// This adapter never sets these fields.
	if got.Sigma != 0 || got.HorizonYears != 0 || got.LiqPriceRatio != 0 || got.Pegged {
		t.Errorf("unexpected market-leg fields set: %+v", got)
	}
}

func TestNotAaveStrategyNoData(t *testing.T) {
	if _, ok := reader(mockReserve{isActive: true}).Risk(aaveID(2)); ok {
		t.Errorf("non-Aave id: ok=true, want false")
	}
}

func TestUnconfiguredBackendNoData(t *testing.T) {
	// nil backend + zero provider => unconfigured => no-data path.
	r := NewAaveChaosReader(nil, common.Address{}, usdcAsset, map[types.StrategyID]bool{aaveID(1): true})
	if _, ok := r.Risk(aaveID(1)); ok {
		t.Errorf("unconfigured reader: ok=true, want false")
	}
}

func TestNilReaderNoData(t *testing.T) {
	var r *AaveChaosReader
	if _, ok := r.Risk(aaveID(1)); ok {
		t.Errorf("nil reader: ok=true, want false")
	}
}

func TestReadErrorDegrades(t *testing.T) {
	base := mockReserve{
		totalAToken: usdc(1_000_000), totalVariableDebt: usdc(500_000),
		decimals: big.NewInt(6), supplyCap: big.NewInt(0), isActive: true,
	}
	mutators := map[string]func(m *mockReserve){
		"data err":   func(m *mockReserve) { m.dataErr = errors.New("x") },
		"cfg err":    func(m *mockReserve) { m.cfgErr = errors.New("x") },
		"caps err":   func(m *mockReserve) { m.capsErr = errors.New("x") },
		"paused err": func(m *mockReserve) { m.pausedErr = errors.New("x") },
	}
	for name, mut := range mutators {
		m := base
		mut(&m)
		if _, ok := reader(m).Risk(aaveID(1)); ok {
			t.Errorf("%s: ok=true, want false (degrade to no-data)", name)
		}
	}
}

// End-to-end: the reader's live signal flows through the real
// GauntletLiteDecider without error and the decider produces a target. This
// adapter contributes the liquidity-haircut and vendor-EL channels; it does NOT
// set the market-price leg (LiqPriceRatio/Sigma), so by design its closed-form PD
// is 0 and a paused reserve does not by itself drop the cap below the owner cap —
// the haircut only bites once another adapter (e.g. a depeg oracle) supplies the
// price barrier. This test pins that documented composition rather than claiming
// a gate the math does not produce.
func TestSignalFlowsThroughBrain(t *testing.T) {
	m := mockReserve{
		totalAToken: usdc(1_000_000), totalVariableDebt: usdc(900_000),
		decimals: big.NewInt(6), supplyCap: big.NewInt(0), isActive: true, paused: true,
	}
	dec := brain.NewGauntletLiteDecider(reader(m), brain.DefaultRiskModel(), nil)
	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: aaveID(1), CapBps: 5000, APY: 0.10},
	}}
	a, err := dec.Decide(nil, state)
	if err != nil {
		t.Fatal(err)
	}
	// Vendor EL (0.02) == SoftThreshold and PD==0, so the cap is not reduced; the
	// owner cap is honored. The point is the live signal was consumed end to end.
	if got := a.Targets[aaveID(1)]; got != 5000 {
		t.Errorf("paused Aave target = %v, want 5000 (haircut needs a price-leg to bite)", got)
	}
}
