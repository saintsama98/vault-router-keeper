package pendle

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/pendleoracle"
	"github.com/vault-router-keeper/internal/bindings/pendlept"
	"github.com/vault-router-keeper/pkg/types"
)

var errUnconfigured = errors.New("pendle: strategy not configured")

// MarketConfig holds the on-chain handles for one PT strategy.
type MarketConfig struct {
	Market   common.Address // Pendle PT/SY market (queried via the oracle)
	PT       common.Address // PrincipalToken (queried for expiry)
	Duration uint32         // TWAP window, seconds
}

// OnChain implements Chain over the abigen bindings against a read-only backend
// (e.g. *ethclient.Client). It is the production wiring; unit tests mock Chain
// directly and never touch this.
type OnChain struct {
	backend bind.ContractCaller
	oracle  common.Address // shared PendlePYLpOracle address
	markets map[types.StrategyID]MarketConfig
}

func NewOnChain(backend bind.ContractCaller, oracle common.Address, markets map[types.StrategyID]MarketConfig) *OnChain {
	return &OnChain{backend: backend, oracle: oracle, markets: markets}
}

func (c *OnChain) OracleReady(id types.StrategyID) (bool, error) {
	mc, ok := c.markets[id]
	if !ok {
		return false, errUnconfigured
	}
	o, err := pendleoracle.NewPendleOracleCaller(c.oracle, c.backend)
	if err != nil {
		return false, err
	}
	st, err := o.GetOracleState(nil, mc.Market, mc.Duration)
	if err != nil {
		return false, err
	}
	return !st.IncreaseCardinalityRequired && st.OldestObservationSatisfied, nil
}

func (c *OnChain) PtToAssetRate(id types.StrategyID) (*big.Int, error) {
	mc, ok := c.markets[id]
	if !ok {
		return nil, errUnconfigured
	}
	o, err := pendleoracle.NewPendleOracleCaller(c.oracle, c.backend)
	if err != nil {
		return nil, err
	}
	return o.GetPtToAssetRate(nil, mc.Market, mc.Duration)
}

func (c *OnChain) Expiry(id types.StrategyID) (*big.Int, error) {
	mc, ok := c.markets[id]
	if !ok {
		return nil, errUnconfigured
	}
	pt, err := pendlept.NewPendlePTCaller(mc.PT, c.backend)
	if err != nil {
		return nil, err
	}
	return pt.Expiry(nil)
}
