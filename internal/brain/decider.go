// Package brain computes the target allocation vector from vault state. This is
// the swappable "intelligence" — the keeper loop never changes when it does.
package brain

import (
	"context"

	"github.com/vault-router-keeper/pkg/types"
)

// Decider maps current vault state to a desired target allocation. A real
// implementation (e.g. marginal-rate water-filling, clamped to caps + idle
// floor, with a risk gate and hysteresis) plugs in here unchanged.
type Decider interface {
	Decide(ctx context.Context, state *types.VaultState) (*types.Allocation, error)
}

// StubDecider returns the current on-chain targets unchanged — it never proposes
// a reallocation. This keeps the keeper "dull": a pure executor of harvest /
// guard / withdraw-fulfillment until a real brain is added.
type StubDecider struct{}

func NewStubDecider() *StubDecider { return &StubDecider{} }

func (StubDecider) Decide(ctx context.Context, state *types.VaultState) (*types.Allocation, error) {
	targets := make(map[types.StrategyID]types.Bps, len(state.Strategies))
	for _, s := range state.Strategies {
		targets[s.ID] = s.TargetBps
	}
	return &types.Allocation{Targets: targets}, nil
}
