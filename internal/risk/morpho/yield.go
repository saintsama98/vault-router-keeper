package morpho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vault-router-keeper/pkg/types"
)

// Doer is the minimal HTTP transport the reader needs (mirrors
// internal/risk/credora.Doer). The real *http.Client satisfies it; tests
// inject a recorder/fixture.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// vaultAPYQuery is the single GraphQL operation this adapter issues. Verified
// against the live API 2026-06-12: state.netApy is the depositor-realized rate
// net of the vault's performance fee (state.apy is gross). chainId is required
// to disambiguate same-address deployments across chains.
const vaultAPYQuery = `query($address: String!, $chainId: Int){ vaultByAddress(address:$address, chainId:$chainId){ state { netApy } } }`

// gqlResponse mirrors exactly the slice of the response we consume.
type gqlResponse struct {
	Data struct {
		VaultByAddress struct {
			State struct {
				NetApy float64 `json:"netApy"`
			} `json:"state"`
		} `json:"vaultByAddress"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// YieldReader implements perceive.YieldProvider for MetaMorpho strategies.
// Each owned StrategyID maps to its vault address; APY() POSTs one GraphQL
// query and returns state.netApy as a decimal fraction. Unowned ids, an empty
// endpoint, transport errors, GraphQL errors, and non-positive rates all
// return ok=false (no-data) — never a fabricated number.
type YieldReader struct {
	doer     Doer
	endpoint string
	chainID  int64
	vaults   map[types.StrategyID]string // id -> ERC-4626 vault address (hex)
}

// NewYieldReader builds the live Morpho net-APY provider. doer defaults to
// http.DefaultClient when nil; an empty endpoint or vault map leaves the
// reader permanently in the no-data path (dormant, same convention as the
// other adapters).
func NewYieldReader(doer Doer, endpoint string, chainID int64, vaults map[types.StrategyID]string) *YieldReader {
	if doer == nil {
		doer = http.DefaultClient
	}
	return &YieldReader{doer: doer, endpoint: endpoint, chainID: chainID, vaults: vaults}
}

// APY returns the vault's live net APY for an owned Morpho strategy.
func (r *YieldReader) APY(id types.StrategyID) (float64, bool) {
	if r == nil || r.endpoint == "" {
		return 0, false
	}
	vault, owned := r.vaults[id]
	if !owned || vault == "" {
		return 0, false
	}

	payload, err := json.Marshal(map[string]any{
		"query": vaultAPYQuery,
		"variables": map[string]any{
			"address": vault,
			"chainId": r.chainID,
		},
	})
	if err != nil {
		return 0, false
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return 0, false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.doer.Do(req)
	if err != nil {
		return 0, false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return 0, false
	}

	var out gqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, false
	}
	if len(out.Errors) > 0 {
		return 0, false
	}
	apy := out.Data.VaultByAddress.State.NetApy
	if apy <= 0 {
		return 0, false
	}
	return apy, true
}

// String aids the keeper's wiring log.
func (r *YieldReader) String() string {
	return fmt.Sprintf("morpho.YieldReader{endpoint:%s chain:%d vaults:%d}", r.endpoint, r.chainID, len(r.vaults))
}
