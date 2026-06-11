package aave

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/aavedata"
	"github.com/vault-router-keeper/pkg/types"
)

// liquidityRateCaller is the narrow on-chain read surface the Aave YieldReader
// needs from the Aave V3 ProtocolDataProvider: just GetReserveData, whose
// LiquidityRate field is the reserve's current supply APR in RAY (1e27) fixed
// point. It is satisfied in production by the abigen binding's
// *aavedata.AaveDataCaller (the anonymous struct mirrors the generated field
// names exactly, so the binding implements it with no adapter) and mocked
// directly in unit tests so they never touch a live chain.
//
// reserveCaller (in chaos.go) is a superset of this surface, so the live
// *aavedata.AaveDataCaller used by AaveChaosReader also satisfies this — but the
// yield reader declares only what it reads.
type liquidityRateCaller interface {
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
}

// rayScale is Aave's RAY fixed-point base (1e27). LiquidityRate is the current
// supply APR expressed in RAY, so the decimal-fraction APR is LiquidityRate/1e27
// (e.g. 5e25 ray -> 0.05 == 5%). This is Aave's documented rate unit, not a
// fabricated coefficient: the rate INPUT is a live on-chain read; only the
// 1e27 divisor (Aave's fixed RAY convention) is a documented default.
var rayScale = new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(27), nil))

// rayToFloat converts a RAY-scaled (1e27) value to a float64 fraction.
func rayToFloat(ray *big.Int) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(ray), rayScale).Float64()
	return f
}

// YieldReader implements perceive.YieldProvider for Aave V3 supply strategies.
// For each owned strategy id it reads the reserve's current LiquidityRate (supply
// APR in RAY) from the live ProtocolDataProvider and returns it as a decimal
// APR fraction (LiquidityRate / 1e27).
//
// It reuses the same aavedata binding/caller pattern as AaveChaosReader: a live
// *aavedata.AaveDataCaller built over the read-only backend, or a unconfigured
// no-data path. Returns ok=false when the id is not an Aave strategy, the
// backend/provider is absent, or the read fails / is non-positive — the honest
// no-data path: the OverlayReader then leaves APY at 0 and the brain water-fills
// by EL alone. This NEVER fabricates a yield: every number traces to a live read.
type YieldReader struct {
	caller liquidityRateCaller // nil => unconfigured backend (no-data path)
	asset  common.Address      // reserve to query == Vault.asset() (USDC)
	isAave map[types.StrategyID]bool
}

// NewYieldReader builds a live Aave supply-APY provider over a read-only backend
// (e.g. *ethclient.Client). provider is the Aave V3 ProtocolDataProvider address;
// asset is the reserve to query (Vault.asset(), i.e. USDC); isAave marks which
// strategy ids this adapter owns.
//
// If backend is nil or provider is the zero address, the reader is unconfigured
// and every APY call returns ok=false (the OverlayReader leaves APY at 0 — the
// honest no-data path, not a stub).
func NewYieldReader(backend bind.ContractCaller, provider, asset common.Address, isAave map[types.StrategyID]bool) *YieldReader {
	r := &YieldReader{asset: asset, isAave: isAave}
	if backend != nil && provider != (common.Address{}) {
		// NewAaveDataCaller only fails if the embedded ABI is malformed, a
		// build-time invariant of the generated binding; treat a failure as
		// unconfigured so we degrade to the no-data path rather than panic.
		if c, err := aavedata.NewAaveDataCaller(provider, backend); err == nil {
			r.caller = c
		}
	}
	return r
}

// newYieldReaderFromCaller is the test seam: it injects a liquidityRateCaller
// directly (a mock), bypassing the live binding construction.
func newYieldReaderFromCaller(caller liquidityRateCaller, asset common.Address, isAave map[types.StrategyID]bool) *YieldReader {
	return &YieldReader{caller: caller, asset: asset, isAave: isAave}
}

// APY returns the live Aave supply APR (LiquidityRate / 1e27) for an owned Aave
// strategy. Returns ok=false when the id is not an Aave strategy, the backend is
// unconfigured, the read fails, or the rate is non-positive.
func (r *YieldReader) APY(id types.StrategyID) (float64, bool) {
	if r == nil || r.caller == nil || !r.isAave[id] {
		return 0, false
	}
	data, err := r.caller.GetReserveData(nil, r.asset)
	if err != nil {
		return 0, false
	}
	if data.LiquidityRate == nil || data.LiquidityRate.Sign() <= 0 {
		return 0, false
	}
	apy := rayToFloat(data.LiquidityRate)
	if apy <= 0 {
		return 0, false
	}
	return apy, true
}
