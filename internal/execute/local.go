package execute

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/vault-router-keeper/internal/bindings/vault"
	ktypes "github.com/vault-router-keeper/pkg/types"
)

// Backend is the subset of bind.ContractBackend a LocalExecutor needs to bind a
// transactor and have go-ethereum estimate gas / nonce / fees. *ethclient.Client
// satisfies it.
type Backend interface {
	bind.ContractBackend
}

// LocalExecutor is the self-hosted signing executor: it holds the curator key in
// process and submits the curator-gated vault calls (setAllocation, rebalance,
// harvestAll, fulfillWithdraw) plus the permissionless guardCheckpoint. It is
// the live counterpart of LogExecutor, selected only when KEEPER_DRY_RUN=false.
//
// It stays dumb plumbing: the brain chose the targets and the validator + the
// on-chain bounded executor both re-check them, so this layer never re-validates
// caps / idle-floor / churn — it just signs and sends.
type LocalExecutor struct {
	vault   *vault.VaultTransactor
	auth    *bind.TransactOpts
	chainID *big.Int
	log     *slog.Logger
}

// NewLocalExecutor binds a transactor to the diamond at vaultAddr and prepares a
// signer from key. chainID must match the connected network (EIP-155 replay
// protection). The curator key is passed in by the caller (read from the env var
// named by cfg.KeyEnv) and never logged.
func NewLocalExecutor(
	backend Backend,
	vaultAddr common.Address,
	key *ecdsa.PrivateKey,
	chainID *big.Int,
	log *slog.Logger,
) (*LocalExecutor, error) {
	t, err := vault.NewVaultTransactor(vaultAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("bind vault transactor: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("keyed transactor: %w", err)
	}
	return &LocalExecutor{vault: t, auth: auth, chainID: chainID, log: log}, nil
}

// Execute signs and submits the transaction for one Action and returns its hash.
// Gas, nonce and fee estimation are left to go-ethereum (auth fields nil).
func (e *LocalExecutor) Execute(ctx context.Context, action ktypes.Action) (string, error) {
	opts := *e.auth // copy so per-call Context never mutates the shared signer
	opts.Context = ctx

	var (
		tx  *types.Transaction
		err error
	)
	switch action.Kind {
	case ktypes.ActionSetAllocation:
		if action.Allocation == nil {
			return "", fmt.Errorf("setAllocation: nil allocation")
		}
		ids, bps := flattenAllocation(action.Allocation)
		tx, err = e.vault.SetAllocation(&opts, ids, bps)
	case ktypes.ActionRebalance:
		tx, err = e.vault.Rebalance(&opts)
	case ktypes.ActionHarvestAll:
		tx, err = e.vault.HarvestAll(&opts)
	case ktypes.ActionFulfillWithdraw:
		if action.WithdrawID == nil {
			return "", fmt.Errorf("fulfillWithdraw: nil withdraw id")
		}
		tx, err = e.vault.FulfillWithdraw(&opts, action.WithdrawID)
	case ktypes.ActionGuardCheckpoint:
		tx, err = e.vault.GuardCheckpoint(&opts)
	default:
		return "", fmt.Errorf("unknown action kind %d", action.Kind)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", action.Kind.String(), err)
	}
	return tx.Hash().Hex(), nil
}

// flattenAllocation turns the brain's target map into the parallel (ids, bps)
// arrays setAllocation expects, sorted by id so the same Allocation always
// produces byte-identical calldata (determinism, PLAN.md §6 invariant 4).
func flattenAllocation(a *ktypes.Allocation) ([][32]byte, []uint16) {
	ids := make([][32]byte, 0, len(a.Targets))
	for id := range a.Targets {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return string(ids[i][:]) < string(ids[j][:])
	})
	bps := make([]uint16, len(ids))
	for i, id := range ids {
		bps[i] = uint16(a.Targets[id])
	}
	return ids, bps
}

// LoadKey parses a hex curator private key (with or without 0x prefix). Kept
// here so main.go can read os.Getenv(cfg.KeyEnv) and hand the parsed key in
// without the key material crossing any other package boundary.
func LoadKey(hexKey string) (*ecdsa.PrivateKey, error) {
	if len(hexKey) >= 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}
	return crypto.HexToECDSA(hexKey)
}
