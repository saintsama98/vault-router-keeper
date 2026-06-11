package credora

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/pkg/types"
)

// Doer is the minimal HTTP transport the reader needs. The real *http.Client
// satisfies it; tests inject a fake that returns recorded fixtures so no live
// endpoint is ever hit.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Reader implements brain.RiskProvider for the Morpho facet. It resolves a
// StrategyID to its Morpho market key, fetches the Credora risk rating over
// GraphQL, and maps that rating to a vendor expected-loss. Any failure (no
// market mapping, transport error, non-200, unparseable rating) degrades to
// "no data" (false) so the brain falls back to the closed-form floor — a broken
// or missing Credora feed can never raise an allocation.
type Reader struct {
	endpoint string
	apiKey   string // optional Bearer token (Credora API is auth-gated)
	doer     Doer
	markets  map[types.StrategyID]string // StrategyID → Morpho market key
}

// NewReader builds a Credora reader. doer defaults to http.DefaultClient when nil.
// An empty endpoint OR an empty markets map means the reader has no live source:
// every Risk call returns ok=false (the honest no-data path; the brain then uses
// its closed-form EL floor). The real Credora-by-RedStone GraphQL API is auth-
// gated, so apiKey (when set) is sent as a Bearer token; pslQuery/pslResponse
// remain operator-reconcilable placeholders until the live schema is supplied.
func NewReader(endpoint, apiKey string, doer Doer, markets map[types.StrategyID]string) *Reader {
	if doer == nil {
		doer = http.DefaultClient
	}
	return &Reader{endpoint: endpoint, apiKey: apiKey, doer: doer, markets: markets}
}

func (r *Reader) Risk(id types.StrategyID) (brain.RiskInputs, bool) {
	if r == nil || r.endpoint == "" {
		return brain.RiskInputs{}, false
	}
	market, ok := r.markets[id]
	if !ok {
		return brain.RiskInputs{}, false
	}
	rating, ok := r.fetchRating(market)
	if !ok {
		return brain.RiskInputs{}, false
	}
	el, ok := ratingToEL(rating)
	if !ok {
		return brain.RiskInputs{}, false
	}
	return brain.RiskInputs{HasVendor: true, VendorEL: el}, true
}

// pslQuery is the GraphQL query for a market's Credora rating.
//
// NOTE (Phase 2 placeholder): the query string and the response shape parsed in
// fetchRating are an ASSUMED schema. Reconcile both against the live Morpho Blue
// / Credora-by-RedStone GraphQL API before production; the fixture test pins the
// shape so a schema change fails loudly.
const pslQuery = `query($market:String!){market(key:$market){credora{rating}}}`

// pslResponse is the assumed GraphQL response envelope (see pslQuery note).
type pslResponse struct {
	Data struct {
		Market struct {
			Credora struct {
				Rating string `json:"rating"`
			} `json:"credora"`
		} `json:"market"`
	} `json:"data"`
}

func (r *Reader) fetchRating(market string) (string, bool) {
	payload, err := json.Marshal(map[string]any{
		"query":     pslQuery,
		"variables": map[string]string{"market": market},
	})
	if err != nil {
		return "", false
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiKey)
	}

	resp, err := r.doer.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var out pslResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", false
	}
	rating := strings.TrimSpace(out.Data.Market.Credora.Rating)
	if rating == "" {
		return "", false
	}
	return rating, true
}

// ratingToEL maps a Credora letter rating (A best → D worst) to a vendor
// expected-loss in [0,1].
//
// PLACEHOLDER CALIBRATION (PLAN.md §7): these are documented stand-ins, not a
// tuned curve. The brain composes this via max() with its own closed-form EL, so
// an over-conservative mapping only ever tightens. D (0.15) sits above the kill
// threshold (0.10) by design, so a worst-rated market is allocated to zero.
func ratingToEL(rating string) (float64, bool) {
	switch strings.ToUpper(rating) {
	case "A":
		return 0.00, true
	case "B":
		return 0.03, true
	case "C":
		return 0.08, true
	case "D":
		return 0.15, true
	default:
		return 0, false
	}
}
