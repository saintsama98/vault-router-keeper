package perceive

import "github.com/vault-router-keeper/pkg/types"

// YieldProvider maps a strategy id to an off-chain APY signal. It mirrors
// brain.RiskProvider so APY is sourced through the same ports/adapters idiom as
// risk. ok=false means "no live yield for this strategy" — the OverlayReader then
// leaves the snapshot's APY untouched (0), and the brain water-fills by EL alone.
//
// APY is never read from the vault diamond (the ChainReader leaves it 0 by
// design); only a YieldProvider populates it, and only from a LIVE feed. No
// adapter fabricates a yield number.
type YieldProvider interface {
	APY(id types.StrategyID) (float64, bool)
}

// CompositeYieldProvider routes each StrategyID to the yield adapter that owns
// it, mirroring internal/risk.CompositeProvider. An id with no registered
// adapter (or a nil router) reports ok=false.
type CompositeYieldProvider struct {
	routes map[types.StrategyID]YieldProvider
}

func NewCompositeYieldProvider(routes map[types.StrategyID]YieldProvider) *CompositeYieldProvider {
	return &CompositeYieldProvider{routes: routes}
}

func (c *CompositeYieldProvider) APY(id types.StrategyID) (float64, bool) {
	if c == nil || c.routes == nil {
		return 0, false
	}
	p, ok := c.routes[id]
	if !ok || p == nil {
		return 0, false
	}
	return p.APY(id)
}
