package perceive

import (
	"context"

	"github.com/vault-router-keeper/pkg/types"
)

// OverlayReader decorates any Reader, overlaying a live off-chain APY onto each
// strategy in the snapshot. It is the ONLY place APY is ever written — the
// underlying ChainReader leaves APY 0 (APY is not on chain), so wrapping it here
// keeps the pure vault reader untouched while still feeding the brain a real
// yield signal where one exists.
//
// For each strategy, if the YieldProvider returns ok, the strategy's APY is set;
// otherwise it is left as-is (0). A nil YieldProvider (or nil receiver inner)
// makes this a pass-through, so wiring the overlay is always safe even when no
// live yield source is configured.
type OverlayReader struct {
	inner  Reader
	yields YieldProvider
}

// NewOverlayReader wraps inner so its snapshots carry live APY from yields. If
// yields is nil the overlay is a pass-through (APY stays 0).
func NewOverlayReader(inner Reader, yields YieldProvider) *OverlayReader {
	return &OverlayReader{inner: inner, yields: yields}
}

func (r *OverlayReader) Snapshot(ctx context.Context) (*types.VaultState, error) {
	state, err := r.inner.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	if r.yields == nil || state == nil {
		return state, err
	}
	for i := range state.Strategies {
		if apy, ok := r.yields.APY(state.Strategies[i].ID); ok {
			state.Strategies[i].APY = apy
		}
	}
	return state, nil
}
