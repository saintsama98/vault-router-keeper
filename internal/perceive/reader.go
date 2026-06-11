// Package perceive reads a consistent snapshot of vault state for the keeper.
package perceive

import (
	"context"
	"math/big"

	"github.com/vault-router-keeper/pkg/types"
)

// Reader produces a single snapshot of vault state. The concrete implementation
// reads AllocatorFacet / GuardFacet / WithdrawQueueFacet views via the generated
// bindings; the stub returns empty state so the loop runs with no chain access.
type Reader interface {
	Snapshot(ctx context.Context) (*types.VaultState, error)
}

// StubReader returns a fixed empty snapshot. Replace with a ChainReader backed
// by internal/chain + internal/vault bindings.
type StubReader struct{}

func NewStubReader() *StubReader { return &StubReader{} }

func (StubReader) Snapshot(ctx context.Context) (*types.VaultState, error) {
	return &types.VaultState{
		TotalAssets:      big.NewInt(0),
		IdleAssets:       big.NewInt(0),
		Strategies:       nil,
		PendingWithdraws: nil,
	}, nil
}
