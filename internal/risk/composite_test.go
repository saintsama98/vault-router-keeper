package risk_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/internal/risk"
	"github.com/vault-router-keeper/internal/risk/credora"
	"github.com/vault-router-keeper/pkg/types"
)

func sid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

// stubProvider is a fixed-EL brain.RiskProvider for routing tests.
type stubProvider struct{ el float64 }

func (s stubProvider) Risk(types.StrategyID) (brain.RiskInputs, bool) {
	return brain.RiskInputs{HasVendor: true, VendorEL: s.el}, true
}

func TestCompositeRouting(t *testing.T) {
	idA, idB, idUnknown := sid(1), sid(2), sid(3)
	comp := risk.NewCompositeProvider(map[types.StrategyID]brain.RiskProvider{
		idA: stubProvider{el: 0.04},
		idB: stubProvider{el: 0.20},
	})

	if ri, ok := comp.Risk(idA); !ok || ri.VendorEL != 0.04 {
		t.Errorf("idA → (%v,%v), want (0.04,true)", ri.VendorEL, ok)
	}
	if ri, ok := comp.Risk(idB); !ok || ri.VendorEL != 0.20 {
		t.Errorf("idB → (%v,%v), want (0.20,true)", ri.VendorEL, ok)
	}
	if _, ok := comp.Risk(idUnknown); ok {
		t.Errorf("unknown id → ok=true, want false")
	}
}

// fakeDoer returns a canned Credora response.
type fakeDoer struct {
	body   string
	status int
}

func (f fakeDoer) Do(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: f.status, Body: io.NopCloser(strings.NewReader(f.body))}, nil
}

// End-to-end: a worst-rated (D) Credora market, routed through the composite,
// must drive that strategy's target to 0 through the real GauntletLiteDecider.
func TestCompositeKillThroughBrain(t *testing.T) {
	id := sid(7)
	cr := credora.NewReader(
		"http://credora.test/graphql",
		"",
		fakeDoer{body: `{"data":{"market":{"credora":{"rating":"D"}}}}`, status: 200},
		map[types.StrategyID]string{id: "morpho-usdc"},
	)
	comp := risk.NewCompositeProvider(map[types.StrategyID]brain.RiskProvider{id: cr})
	dec := brain.NewGauntletLiteDecider(comp, brain.DefaultRiskModel(), nil)

	state := &types.VaultState{Strategies: []types.StrategyState{
		{ID: id, CapBps: 5000, APY: 0.10}, // attractive yield, but rated D
	}}
	a, err := dec.Decide(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Targets[id]; got != 0 {
		t.Errorf("D-rated strategy target = %v, want 0 (vendor EL 0.15 ≥ kill 0.10)", got)
	}
}
