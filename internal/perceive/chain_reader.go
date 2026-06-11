package perceive

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/vault"
	"github.com/vault-router-keeper/pkg/types"
)

// ChainReader reads a live VaultState from the diamond via the generated vault
// bindings. Because the vault is an EIP-2535 diamond, every facet answers at one
// address, so a single VaultCaller covers the Allocator / Guard / WithdrawQueue
// and ERC-4626 views.
//
// Reads are issued at "latest" and are NOT atomic across calls: a rebalance
// landing mid-snapshot could yield a slightly inconsistent view. For a strict
// snapshot, pin every call to one block (multicall / batched eth_call at a fixed
// BlockNumber) — deferred infra (PLAN.md §4 internal/chain multicall). The
// keeper tolerates minor drift because the validator re-checks bounds and the
// vault rejects any out-of-bounds allocation on-chain.
type ChainReader struct {
	caller *vault.VaultCaller
}

// NewChainReader binds a read-only caller to the diamond at vaultAddr. backend
// is any bind.ContractCaller (e.g. *ethclient.Client).
func NewChainReader(backend bind.ContractCaller, vaultAddr common.Address) (*ChainReader, error) {
	c, err := vault.NewVaultCaller(vaultAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("bind vault caller: %w", err)
	}
	return &ChainReader{caller: c}, nil
}

// Snapshot reads one consistent-enough view of vault state. APY is left zero —
// it is an off-chain signal a RiskProvider/feed supplies, never read from chain.
func (r *ChainReader) Snapshot(ctx context.Context) (*types.VaultState, error) {
	opts := &bind.CallOpts{Context: ctx}

	total, err := r.caller.TotalAssets(opts)
	if err != nil {
		return nil, fmt.Errorf("totalAssets: %w", err)
	}
	idle, err := r.caller.IdleAssets(opts)
	if err != nil {
		return nil, fmt.Errorf("idleAssets: %w", err)
	}
	idleReserve, err := r.caller.IdleReserveBps(opts)
	if err != nil {
		return nil, fmt.Errorf("idleReserveBps: %w", err)
	}
	maxDelta, err := r.caller.MaxRebalanceDelta(opts)
	if err != nil {
		return nil, fmt.Errorf("maxRebalanceDelta: %w", err)
	}
	paused, err := r.caller.Paused(opts)
	if err != nil {
		return nil, fmt.Errorf("paused: %w", err)
	}

	strategies, err := r.readStrategies(opts)
	if err != nil {
		return nil, err
	}
	pending, err := r.readPendingWithdraws(opts)
	if err != nil {
		return nil, err
	}

	return &types.VaultState{
		TotalAssets:          total,
		IdleAssets:           idle,
		IdleReserveBps:       types.Bps(idleReserve),
		MaxRebalanceDeltaBps: types.Bps(maxDelta),
		Paused:               paused,
		Strategies:           strategies,
		PendingWithdraws:     pending,
	}, nil
}

func (r *ChainReader) readStrategies(opts *bind.CallOpts) ([]types.StrategyState, error) {
	ids, err := r.caller.Strategies(opts)
	if err != nil {
		return nil, fmt.Errorf("strategies: %w", err)
	}
	out := make([]types.StrategyState, 0, len(ids))
	for _, id := range ids {
		target, err := r.caller.TargetAllocation(opts, id)
		if err != nil {
			return nil, fmt.Errorf("targetAllocation(%x): %w", id[:4], err)
		}
		assets, err := r.caller.StrategyTotalAssets(opts, id)
		if err != nil {
			return nil, fmt.Errorf("strategyTotalAssets(%x): %w", id[:4], err)
		}
		capBps, err := r.caller.StrategyCap(opts, id)
		if err != nil {
			return nil, fmt.Errorf("strategyCap(%x): %w", id[:4], err)
		}
		quarantined, err := r.caller.IsQuarantined(opts, id)
		if err != nil {
			return nil, fmt.Errorf("isQuarantined(%x): %w", id[:4], err)
		}
		out = append(out, types.StrategyState{
			ID:            types.StrategyID(id),
			TargetBps:     types.Bps(target),
			CurrentAssets: assets,
			CapBps:        types.Bps(capBps),
			Quarantined:   quarantined,
			APY:           0, // off-chain signal; not on chain
		})
	}
	return out, nil
}

// readPendingWithdraws enumerates unfulfilled async-queue exits. Request ids run
// [0, nextWithdrawRequestId); a shares==0 slot is empty / fulfilled / cancelled.
// We short-circuit when nothing is escrowed so the common (empty-queue) case
// costs a single call.
func (r *ChainReader) readPendingWithdraws(opts *bind.CallOpts) ([]types.WithdrawRequest, error) {
	totalPending, err := r.caller.PendingWithdrawShares(opts)
	if err != nil {
		return nil, fmt.Errorf("pendingWithdrawShares: %w", err)
	}
	if totalPending.Sign() == 0 {
		return nil, nil
	}
	next, err := r.caller.NextWithdrawRequestId(opts)
	if err != nil {
		return nil, fmt.Errorf("nextWithdrawRequestId: %w", err)
	}

	var out []types.WithdrawRequest
	for i := new(big.Int); i.Cmp(next) < 0; i.Add(i, big.NewInt(1)) {
		id := new(big.Int).Set(i)
		req, err := r.caller.WithdrawRequest(opts, id)
		if err != nil {
			return nil, fmt.Errorf("withdrawRequest(%s): %w", id, err)
		}
		if req.Shares.Sign() == 0 {
			continue
		}
		out = append(out, types.WithdrawRequest{ID: id, Shares: req.Shares})
	}
	return out, nil
}
