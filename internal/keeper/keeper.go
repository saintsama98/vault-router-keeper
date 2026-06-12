// Package keeper wires perceive -> brain -> trigger -> execute into one poll
// loop. It owns NO allocation logic; planning is mechanical (the brain already
// decided the targets).
package keeper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/internal/execute"
	"github.com/vault-router-keeper/internal/perceive"
	"github.com/vault-router-keeper/internal/trigger"
	"github.com/vault-router-keeper/internal/validate"
	"github.com/vault-router-keeper/pkg/types"
)

// Task names used with the Scheduler.
const (
	taskGuard     = "guard"
	taskHarvest   = "harvest"
	taskRebalance = "rebalance"
)

// Keeper polls vault state, asks the brain for a target allocation, plans the
// concrete actions, and hands each to the executor.
type Keeper struct {
	reader    perceive.Reader
	decider   brain.Decider
	validator *validate.Validator
	sched     trigger.Scheduler
	exec      execute.Executor
	pollEvery time.Duration
	log       *slog.Logger
}

func New(
	r perceive.Reader,
	d brain.Decider,
	v *validate.Validator,
	s trigger.Scheduler,
	e execute.Executor,
	pollEvery time.Duration,
	log *slog.Logger,
) *Keeper {
	return &Keeper{reader: r, decider: d, validator: v, sched: s, exec: e, pollEvery: pollEvery, log: log}
}

// Run polls until ctx is cancelled.
func (k *Keeper) Run(ctx context.Context) error {
	t := time.NewTicker(k.pollEvery)
	defer t.Stop()
	k.log.Info("keeper started", "poll", k.pollEvery.String())
	for {
		if err := k.tick(ctx); err != nil {
			k.log.Error("tick failed", "err", err)
		}
		select {
		case <-ctx.Done():
			k.log.Info("keeper stopping")
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (k *Keeper) tick(ctx context.Context) error {
	now := time.Now()

	state, err := k.reader.Snapshot(ctx)
	if err != nil {
		return err
	}
	alloc, err := k.decider.Decide(ctx, state)
	if err != nil {
		return err
	}

	// Safety stage: re-check every hard bound off-chain. On failure, fall back to
	// the current on-chain targets (a no-op reallocation) so the loop never pushes
	// an out-of-bounds allocation; guard/harvest/withdraw still proceed.
	if res := k.validator.Check(state, alloc); !res.OK {
		k.log.Error("allocation rejected by validator; falling back to current targets",
			"violations", res.Violations)
		alloc = currentTargets(state)
	}

	for _, a := range k.plan(state, alloc, now) {
		hash, err := k.exec.Execute(ctx, a)
		if err != nil {
			k.log.Error("execute failed", "kind", a.Kind.String(), "err", err)
			continue
		}
		k.log.Info("submitted", "kind", a.Kind.String(), "tx", hash)
	}
	return nil
}

// plan turns (state, desired allocation, cadence) into the actions to fire this
// tick. Mechanical only — the brain already chose the targets.
func (k *Keeper) plan(state *types.VaultState, alloc *types.Allocation, now time.Time) []types.Action {
	var actions []types.Action

	// Permissionless NAV checkpoint — safe even when paused / no curator key.
	if k.sched.Due(taskGuard, now) {
		actions = append(actions, types.Action{Kind: types.ActionGuardCheckpoint})
		k.sched.Mark(taskGuard, now)
	}

	// Curator-gated actions require an unpaused vault.
	if state.Paused {
		k.log.Warn("vault paused; skipping curator actions")
		return actions
	}

	// Reallocation: only if the brain's targets differ from on-chain, throttled.
	if k.sched.Due(taskRebalance, now) && allocationDiffers(state, alloc) {
		actions = append(actions,
			types.Action{Kind: types.ActionSetAllocation, Allocation: alloc},
			types.Action{Kind: types.ActionRebalance},
		)
		k.sched.Mark(taskRebalance, now)
	}

	// Harvest on cadence.
	if k.sched.Due(taskHarvest, now) {
		actions = append(actions, types.Action{Kind: types.ActionHarvestAll})
		k.sched.Mark(taskHarvest, now)
	}

	// Fulfill pending async withdrawals. The idle-liquidity check is enforced
	// on-chain; a richer planner would pre-check idle and rebalance liquidity
	// out first so the fulfill cannot revert.
	for _, w := range state.PendingWithdraws {
		actions = append(actions, types.Action{Kind: types.ActionFulfillWithdraw, WithdrawID: w.ID})
	}

	return actions
}

// currentTargets is the no-op allocation: every strategy keeps its on-chain
// target. Used as the safe fallback when validation rejects the brain's output.
func currentTargets(state *types.VaultState) *types.Allocation {
	targets := make(map[types.StrategyID]types.Bps, len(state.Strategies))
	for _, s := range state.Strategies {
		targets[s.ID] = s.TargetBps
	}
	return &types.Allocation{Targets: targets}
}

func allocationDiffers(state *types.VaultState, alloc *types.Allocation) bool {
	for _, s := range state.Strategies {
		if alloc.Targets[s.ID] != s.TargetBps {
			return true
		}
	}
	return false
}
