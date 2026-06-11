package aave

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/aavedata"
	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// reserveCaller is the narrow on-chain read surface AaveChaosReader needs from
// the Aave V3 ProtocolDataProvider (a.k.a. AaveProtocolDataProvider). It is
// satisfied in production by the abigen binding's *AaveDataCaller (see
// internal/bindings/aavedata) and mocked directly in unit tests so they never
// touch a live chain.
//
// The names mirror the generated binding exactly (anonymous structs with the
// abigen-derived field names) so *aavedata.AaveDataCaller implements this
// interface with no adapter.
type reserveCaller interface {
	GetReserveData(opts *bind.CallOpts, asset common.Address) (struct {
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
	}, error)

	GetReserveConfigurationData(opts *bind.CallOpts, asset common.Address) (struct {
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
	}, error)

	GetReserveCaps(opts *bind.CallOpts, asset common.Address) (struct {
		BorrowCap *big.Int
		SupplyCap *big.Int
	}, error)

	GetPaused(opts *bind.CallOpts, asset common.Address) (bool, error)
}

// Documented-default curve coefficients (CALIBRATION TODO).
//
// Every INPUT these coefficients are applied to is a LIVE on-chain read
// (utilization derived from getReserveData debt/aToken totals, isActive/isFrozen
// from getReserveConfigurationData, isPaused from getPaused, supplyCap from
// getReserveCaps). Only the shape of the input->risk mapping is a default; the
// data is never fabricated. Tune these against historical Aave V3 USDC
// withdrawal-stress / utilization data before relying on them in production.
const (
	// Step 1 hard-stop haircuts.
	haircutPaused      = 1.00 // paused reserve: capital cannot exit this epoch.
	haircutWindingDown = 0.50 // inactive/frozen: reserve winding down, floor.

	// Step 2 utilization->liquidity-haircut piecewise-linear breakpoints.
	utilHaircutLow  = 0.90 // <= this: no liquidity haircut from utilization.
	utilHaircutHigh = 0.99 // > this: near-100% util, max utilization haircut.
	haircutUtilMid  = 0.50 // haircut at utilHaircutHigh (ramp top of mid band).
	haircutUtilTop  = 0.80 // haircut above utilHaircutHigh.

	// Step 4 vendor-EL util->EL curve.
	vendorELBase   = 0.0005 // 5 bps Aave smart-contract / baseline risk.
	vendorELU0     = 0.80   // utilization above which modeled EL starts ramping.
	vendorELK      = 0.05   // EL slope per unit utilization above u0.
	vendorELCap    = 0.05   // 5% modeled-EL ceiling.
	vendorELFrozen = 0.02   // additional EL floor when paused/frozen.
)

// AaveChaosReader is a LIVE on-chain brain.RiskProvider for Aave V3 supply
// strategies. It reads the reserve's utilization, active/frozen/paused state and
// supply cap from the Aave ProtocolDataProvider and maps them to RiskInputs via
// the documented-default curves above. Every signal is a live read; ok=false is
// returned only when the strategy is not an Aave strategy or the backend/config
// is absent (the honest no-data path: brain then uses its closed-form EL floor).
//
// It is the live successor to ConfiguredReader (operator-set profiles), which is
// retained as a documented fallback for environments without a data-provider
// address.
type AaveChaosReader struct {
	caller   reserveCaller  // nil => unconfigured backend (no-data path)
	asset    common.Address // reserve to query == Vault.asset() (USDC)
	provider common.Address // ProtocolDataProvider address (zero => unconfigured)
	isAave   map[types.StrategyID]bool
}

// NewAaveChaosReader builds a live reader over a read-only backend (e.g.
// *ethclient.Client). provider is the Aave V3 ProtocolDataProvider address;
// asset is the reserve to query (Vault.asset(), i.e. USDC); isAave marks which
// strategy ids this adapter owns.
//
// If backend is nil or provider is the zero address, the reader is "unconfigured"
// and every Risk call returns ok=false (brain falls back to its closed-form EL
// floor — the honest no-data path, not a stub).
func NewAaveChaosReader(backend bind.ContractCaller, provider, asset common.Address, isAave map[types.StrategyID]bool) *AaveChaosReader {
	r := &AaveChaosReader{asset: asset, provider: provider, isAave: isAave}
	if backend != nil && provider != (common.Address{}) {
		// NewAaveDataCaller only fails if the embedded ABI is malformed, which
		// is a build-time invariant of the generated binding; treat a failure as
		// "unconfigured" so we degrade to the no-data path rather than panic.
		if c, err := aavedata.NewAaveDataCaller(provider, backend); err == nil {
			r.caller = c
		}
	}
	return r
}

// newAaveChaosReaderFromCaller is the test seam: it injects a reserveCaller
// directly (a mock), bypassing the live binding construction.
func newAaveChaosReaderFromCaller(caller reserveCaller, asset common.Address, isAave map[types.StrategyID]bool) *AaveChaosReader {
	return &AaveChaosReader{caller: caller, asset: asset, isAave: isAave}
}

// Risk reads the live reserve state for an Aave strategy and maps it to
// RiskInputs. Returns ok=false when the id is not an Aave strategy, the backend
// is unconfigured, or any on-chain read fails (degrade to closed-form floor).
func (r *AaveChaosReader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if r == nil || r.caller == nil || !r.isAave[id] {
		return brain.RiskInputs{}, false
	}

	data, err := r.caller.GetReserveData(nil, r.asset)
	if err != nil {
		return brain.RiskInputs{}, false
	}
	cfg, err := r.caller.GetReserveConfigurationData(nil, r.asset)
	if err != nil {
		return brain.RiskInputs{}, false
	}
	caps, err := r.caller.GetReserveCaps(nil, r.asset)
	if err != nil {
		return brain.RiskInputs{}, false
	}
	paused, err := r.caller.GetPaused(nil, r.asset)
	if err != nil {
		return brain.RiskInputs{}, false
	}

	var out brain.RiskInputs

	// Step 2 input: utilization = totalDebt / totalAToken (both already in token
	// decimals, so the ratio is unitless). Guard totalAToken == 0 => util = 0.
	util := utilization(data.TotalStableDebt, data.TotalVariableDebt, data.TotalAToken)

	// Step 1 — hard stops on the liquidity haircut.
	if paused {
		out.LiquidityHaircut = haircutPaused // 1.0: capital cannot exit.
	}
	if !cfg.IsActive || cfg.IsFrozen {
		if haircutWindingDown > out.LiquidityHaircut {
			out.LiquidityHaircut = haircutWindingDown
		}
		out.DepegWatch = true
	}

	// Step 2 — utilization -> liquidity stress (composed via max with step 1).
	if hu := utilHaircut(util); hu > out.LiquidityHaircut {
		out.LiquidityHaircut = hu
	}

	// Step 3 — supply-cap headroom (informational liquidity ceiling).
	if caps.SupplyCap != nil && caps.SupplyCap.Sign() > 0 {
		out.HasLiquidityCap = true
		out.LiquidityCapBps = supplyHeadroomBps(caps.SupplyCap, cfg.Decimals, data.TotalAToken)
	}

	// Step 4 — vendor EL channel: smooth util->EL curve.
	out.HasVendor = true
	out.VendorEL = vendorEL(util)
	if paused || !cfg.IsActive || cfg.IsFrozen {
		if vendorELFrozen > out.VendorEL {
			out.VendorEL = vendorELFrozen
		}
	}

	return out, true
}

// utilization computes (stableDebt + variableDebt) / totalAToken as a float in
// token-decimal units (the ratio is unitless). totalAToken == 0 (or nil) yields
// 0 — an empty reserve has no withdrawal contention.
func utilization(stableDebt, variableDebt, totalAToken *big.Int) float64 {
	if totalAToken == nil || totalAToken.Sign() <= 0 {
		return 0
	}
	debt := new(big.Float)
	if stableDebt != nil {
		debt.Add(debt, new(big.Float).SetInt(stableDebt))
	}
	if variableDebt != nil {
		debt.Add(debt, new(big.Float).SetInt(variableDebt))
	}
	ratio := new(big.Float).Quo(debt, new(big.Float).SetInt(totalAToken))
	u, _ := ratio.Float64()
	if u < 0 {
		return 0
	}
	return u
}

// utilHaircut is the documented-default piecewise-linear utilization->haircut
// curve (CALIBRATION TODO): flat 0 up to utilHaircutLow, a linear ramp to
// haircutUtilMid across (utilHaircutLow, utilHaircutHigh], then haircutUtilTop
// above utilHaircutHigh.
func utilHaircut(util float64) float64 {
	switch {
	case util <= utilHaircutLow:
		return 0
	case util <= utilHaircutHigh:
		frac := (util - utilHaircutLow) / (utilHaircutHigh - utilHaircutLow)
		return frac * haircutUtilMid
	default:
		return haircutUtilTop
	}
}

// supplyHeadroomBps returns the remaining deposit headroom under the supply cap
// in bps: floor(10000 * (cap*10^decimals - totalAToken) / (cap*10^decimals)),
// clamped to [0,10000]. The supply cap is a whole-token figure, so it is scaled
// by 10^decimals to compare against totalAToken (in token base units). Near-zero
// headroom signals new deposits will revert.
func supplyHeadroomBps(supplyCap, decimals, totalAToken *big.Int) types.Bps {
	if supplyCap == nil || supplyCap.Sign() <= 0 {
		return 0
	}
	dec := int64(6)
	if decimals != nil && decimals.IsInt64() {
		dec = decimals.Int64()
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(dec), nil)
	capBase := new(big.Int).Mul(supplyCap, scale) // cap in token base units
	if capBase.Sign() <= 0 {
		return 0
	}
	supplied := big.NewInt(0)
	if totalAToken != nil {
		supplied = totalAToken
	}
	headroom := new(big.Int).Sub(capBase, supplied)
	if headroom.Sign() <= 0 {
		return 0 // at/over cap: no headroom, deposits will revert.
	}
	// bps = 10000 * headroom / capBase
	bps := new(big.Int).Mul(headroom, big.NewInt(10000))
	bps.Quo(bps, capBase)
	if bps.Cmp(big.NewInt(10000)) > 0 {
		return 10000
	}
	return types.Bps(bps.Int64())
}

// vendorEL is the documented-default smooth util->EL curve (CALIBRATION TODO):
// clamp(base + k*max(0, util-u0), 0, cap). It raises modeled expected loss as the
// reserve approaches illiquidity, starting from the Aave baseline contract risk.
func vendorEL(util float64) float64 {
	el := vendorELBase + vendorELK*math.Max(0, util-vendorELU0)
	if el < 0 {
		return 0
	}
	if el > vendorELCap {
		return vendorELCap
	}
	return el
}
