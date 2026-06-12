package morpho

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/vault-router-keeper/pkg/types"
)

var (
	morphoID = types.StrategyID{'m', 'o', 'r', 'p', 'h', 'o'}
	otherID  = types.StrategyID{'o', 't', 'h', 'e', 'r'}
)

const steakhouse = "0xBEEF01735c132Ada46AA9aA4c54623cAA92A64CB"

// fixtureDoer replays a canned HTTP response and records the request payload.
type fixtureDoer struct {
	status  int
	body    string
	err     error
	lastReq map[string]any
}

func (f *fixtureDoer) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, &f.lastReq)
	}
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(bytes.NewBufferString(f.body)),
	}, nil
}

// liveFixture is the exact response shape recorded from the real API on
// 2026-06-12 (Steakhouse USDC, chain 1) — the anti-hallucination oracle.
const liveFixture = `{"data":{"vaultByAddress":{"address":"0xBEEF01735c132Ada46AA9aA4c54623cAA92A64CB",` +
	`"state":{"apy":0.043727269079311945,"netApy":0.04149617540233034}}},` +
	`"extensions":{"complexity":50,"maximumComplexity":1000000}}`

func TestAPY(t *testing.T) {
	vaults := map[types.StrategyID]string{morphoID: steakhouse}

	tests := []struct {
		name     string
		doer     *fixtureDoer
		endpoint string
		id       types.StrategyID
		wantAPY  float64
		wantOK   bool
	}{
		{
			name:     "live fixture returns netApy",
			doer:     &fixtureDoer{status: 200, body: liveFixture},
			endpoint: "https://blue-api.morpho.org/graphql",
			id:       morphoID,
			wantAPY:  0.04149617540233034, // netApy, NOT the gross 0.0437
			wantOK:   true,
		},
		{
			name:     "unowned id is no-data",
			doer:     &fixtureDoer{status: 200, body: liveFixture},
			endpoint: "https://blue-api.morpho.org/graphql",
			id:       otherID,
			wantOK:   false,
		},
		{
			name:     "empty endpoint stays dormant",
			doer:     &fixtureDoer{status: 200, body: liveFixture},
			endpoint: "",
			id:       morphoID,
			wantOK:   false,
		},
		{
			name:     "graphql error degrades to no-data",
			doer:     &fixtureDoer{status: 200, body: `{"errors":[{"message":"boom"}]}`},
			endpoint: "https://blue-api.morpho.org/graphql",
			id:       morphoID,
			wantOK:   false,
		},
		{
			name:     "http 500 degrades to no-data",
			doer:     &fixtureDoer{status: 500, body: `{}`},
			endpoint: "https://blue-api.morpho.org/graphql",
			id:       morphoID,
			wantOK:   false,
		},
		{
			name:     "zero netApy is no-data, never a fabricated 0%",
			doer:     &fixtureDoer{status: 200, body: `{"data":{"vaultByAddress":{"state":{"netApy":0}}}}`},
			endpoint: "https://blue-api.morpho.org/graphql",
			id:       morphoID,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewYieldReader(tt.doer, tt.endpoint, 1, vaults)
			apy, ok := r.APY(tt.id)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if diff := apy - tt.wantAPY; diff > 1e-12 || diff < -1e-12 {
				t.Fatalf("apy = %v, want %v", apy, tt.wantAPY)
			}
			// The request must carry the vault address and chain id verbatim.
			vars, _ := tt.doer.lastReq["variables"].(map[string]any)
			if vars["address"] != steakhouse {
				t.Fatalf("request address = %v, want %s", vars["address"], steakhouse)
			}
			if vars["chainId"] != float64(1) {
				t.Fatalf("request chainId = %v, want 1", vars["chainId"])
			}
		})
	}
}

func TestNilReaderIsNoData(t *testing.T) {
	var r *YieldReader
	if _, ok := r.APY(morphoID); ok {
		t.Fatal("nil reader must be no-data")
	}
}
