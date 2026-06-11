package risk

import (
	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// CompositeProvider routes each StrategyID to the facet adapter that owns it,
// then delegates the EL lookup to that adapter. It implements brain.RiskProvider,
// so the brain sees one provider regardless of how many heterogeneous feeds sit
// behind it. A strategy with no registered adapter reports "no data" (false),
// which the brain treats as the closed-form floor (UnknownEL) — never a reason
// to raise a cap.
type CompositeProvider struct {
	routes map[types.StrategyID]brain.RiskProvider
}

// NewCompositeProvider builds a router from a StrategyID → adapter map. Point
// several ids at the same adapter instance to route a whole facet through it.
func NewCompositeProvider(routes map[types.StrategyID]brain.RiskProvider) *CompositeProvider {
	return &CompositeProvider{routes: routes}
}

func (c *CompositeProvider) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if c == nil || c.routes == nil {
		return brain.RiskInputs{}, false
	}
	p, ok := c.routes[id]
	if !ok || p == nil {
		return brain.RiskInputs{}, false
	}
	return p.Risk(id)
}
