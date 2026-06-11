package perceive

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/vault-router-keeper/internal/bindings/vault"
	"github.com/vault-router-keeper/pkg/types"
)

// mockCaller is a deterministic bind.ContractCaller. It decodes the ABI
// selector (and any bytes32 strategy-id argument) out of each eth_call and
// returns canned, ABI-encoded output packed with the real vault ABI. This
// exercises the generated *vault.VaultCaller end-to-end (pack request ->
// CallContract -> unpack response) without any network or chain.
//
// callCounts records how many times each Solidity method was invoked so the
// test can assert the withdraw-queue short-circuit (withdrawRequest must NOT
// be called when pendingWithdrawShares == 0).
type mockCaller struct {
	abi *abi.ABI

	totalAssets           *big.Int
	idleAssets            *big.Int
	idleReserveBps        uint16
	maxRebalanceDelta     uint16
	paused                bool
	strategyIDs           [][32]byte
	pendingWithdrawShares *big.Int
	nextWithdrawRequestID *big.Int

	// per-strategy canned values, keyed by bytes32 id.
	targetByID      map[[32]byte]uint16
	capByID         map[[32]byte]uint16
	assetsByID      map[[32]byte]*big.Int
	quarantinedByID map[[32]byte]bool

	// per-id withdraw requests for the queue scan.
	withdrawByID map[string]vault.LibWithdrawQueueWithdrawRequest

	callCounts map[string]int
}

// methodBySelector resolves the 4-byte selector at the head of calldata back to
// the ABI method (and thus its name + output args).
func (m *mockCaller) methodBySelector(data []byte) (*abi.Method, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("calldata too short: %d bytes", len(data))
	}
	return m.abi.MethodById(data[:4])
}

// CodeAt is only consulted by the binding when CallContract returns empty
// output; returning a non-empty byte slice keeps the binding from raising
// ErrNoCode. We never want the binding to treat our canned outputs as "no
// contract here".
func (m *mockCaller) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return []byte{0x60, 0x00}, nil
}

func (m *mockCaller) CallContract(_ context.Context, call ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	method, err := m.methodBySelector(call.Data)
	if err != nil {
		return nil, err
	}
	if m.callCounts == nil {
		m.callCounts = map[string]int{}
	}
	m.callCounts[method.Name]++

	// Decode the (optional) single bytes32 / uint256 argument.
	var args []interface{}
	if len(method.Inputs) > 0 {
		args, err = method.Inputs.Unpack(call.Data[4:])
		if err != nil {
			return nil, err
		}
	}

	switch method.Name {
	case "totalAssets":
		return method.Outputs.Pack(m.totalAssets)
	case "idleAssets":
		return method.Outputs.Pack(m.idleAssets)
	case "idleReserveBps":
		return method.Outputs.Pack(m.idleReserveBps)
	case "maxRebalanceDelta":
		return method.Outputs.Pack(m.maxRebalanceDelta)
	case "paused":
		return method.Outputs.Pack(m.paused)
	case "strategies":
		return method.Outputs.Pack(m.strategyIDs)
	case "pendingWithdrawShares":
		return method.Outputs.Pack(m.pendingWithdrawShares)
	case "nextWithdrawRequestId":
		return method.Outputs.Pack(m.nextWithdrawRequestID)
	case "targetAllocation":
		id := args[0].([32]byte)
		return method.Outputs.Pack(m.targetByID[id])
	case "strategyCap":
		id := args[0].([32]byte)
		return method.Outputs.Pack(m.capByID[id])
	case "strategyTotalAssets":
		id := args[0].([32]byte)
		return method.Outputs.Pack(m.assetsByID[id])
	case "isQuarantined":
		id := args[0].([32]byte)
		return method.Outputs.Pack(m.quarantinedByID[id])
	case "withdrawRequest":
		id := args[0].(*big.Int)
		req := m.withdrawByID[id.String()]
		if req.Owner == (common.Address{}) && req.Receiver == (common.Address{}) && req.Shares == nil {
			// empty/zero slot
			return method.Outputs.Pack(vault.LibWithdrawQueueWithdrawRequest{Shares: big.NewInt(0)})
		}
		return method.Outputs.Pack(req)
	default:
		return nil, fmt.Errorf("unexpected method call: %s", method.Name)
	}
}

func mustABI(t *testing.T) *abi.ABI {
	t.Helper()
	parsed, err := vault.VaultMetaData.GetAbi()
	if err != nil {
		t.Fatalf("parse vault ABI: %v", err)
	}
	return parsed
}

func id32(b byte) [32]byte {
	var out [32]byte
	out[0] = b
	return out
}

func TestChainReader_Snapshot(t *testing.T) {
	vaultAddr := common.HexToAddress("0x000000000000000000000000000000000000dEaD")

	idA := id32(0xAA)
	idB := id32(0xBB)

	tests := []struct {
		name  string
		mock  *mockCaller
		check func(t *testing.T, mc *mockCaller, st *types.VaultState)
	}{
		{
			name: "two-strategy snapshot maps ids/targets/caps/quarantine and scalars; empty queue short-circuits",
			mock: &mockCaller{
				totalAssets:           big.NewInt(1_000_000),
				idleAssets:            big.NewInt(250_000),
				idleReserveBps:        500,
				maxRebalanceDelta:     1500,
				paused:                false,
				strategyIDs:           [][32]byte{idA, idB},
				pendingWithdrawShares: big.NewInt(0), // empty queue -> short-circuit
				nextWithdrawRequestID: big.NewInt(7),
				targetByID:            map[[32]byte]uint16{idA: 6000, idB: 4000},
				capByID:               map[[32]byte]uint16{idA: 7000, idB: 5000},
				assetsByID:            map[[32]byte]*big.Int{idA: big.NewInt(450_000), idB: big.NewInt(300_000)},
				quarantinedByID:       map[[32]byte]bool{idA: false, idB: true},
			},
			check: func(t *testing.T, mc *mockCaller, st *types.VaultState) {
				if st.TotalAssets.Cmp(big.NewInt(1_000_000)) != 0 {
					t.Errorf("TotalAssets = %s, want 1000000", st.TotalAssets)
				}
				if st.IdleAssets.Cmp(big.NewInt(250_000)) != 0 {
					t.Errorf("IdleAssets = %s, want 250000", st.IdleAssets)
				}
				if st.IdleReserveBps != types.Bps(500) {
					t.Errorf("IdleReserveBps = %d, want 500", st.IdleReserveBps)
				}
				if st.MaxRebalanceDeltaBps != types.Bps(1500) {
					t.Errorf("MaxRebalanceDeltaBps = %d, want 1500", st.MaxRebalanceDeltaBps)
				}
				if st.Paused {
					t.Errorf("Paused = true, want false")
				}

				if len(st.Strategies) != 2 {
					t.Fatalf("len(Strategies) = %d, want 2", len(st.Strategies))
				}
				// order must match the strategies() return order.
				wantIDs := []types.StrategyID{types.StrategyID(idA), types.StrategyID(idB)}
				wantTargets := []types.Bps{6000, 4000}
				wantCaps := []types.Bps{7000, 5000}
				wantAssets := []*big.Int{big.NewInt(450_000), big.NewInt(300_000)}
				wantQuar := []bool{false, true}
				for i, s := range st.Strategies {
					if s.ID != wantIDs[i] {
						t.Errorf("Strategies[%d].ID = %x, want %x", i, s.ID, wantIDs[i])
					}
					if s.TargetBps != wantTargets[i] {
						t.Errorf("Strategies[%d].TargetBps = %d, want %d", i, s.TargetBps, wantTargets[i])
					}
					if s.CapBps != wantCaps[i] {
						t.Errorf("Strategies[%d].CapBps = %d, want %d", i, s.CapBps, wantCaps[i])
					}
					if s.CurrentAssets.Cmp(wantAssets[i]) != 0 {
						t.Errorf("Strategies[%d].CurrentAssets = %s, want %s", i, s.CurrentAssets, wantAssets[i])
					}
					if s.Quarantined != wantQuar[i] {
						t.Errorf("Strategies[%d].Quarantined = %v, want %v", i, s.Quarantined, wantQuar[i])
					}
					if s.APY != 0 {
						t.Errorf("Strategies[%d].APY = %v, want 0 (off-chain)", i, s.APY)
					}
				}

				// short-circuit: empty queue means no per-id scan at all.
				if len(st.PendingWithdraws) != 0 {
					t.Errorf("PendingWithdraws = %v, want empty", st.PendingWithdraws)
				}
				if mc.callCounts["withdrawRequest"] != 0 {
					t.Errorf("withdrawRequest called %d times, want 0 (short-circuit)", mc.callCounts["withdrawRequest"])
				}
				if mc.callCounts["nextWithdrawRequestId"] != 0 {
					t.Errorf("nextWithdrawRequestId called %d times, want 0 (short-circuit)", mc.callCounts["nextWithdrawRequestId"])
				}
				if mc.callCounts["pendingWithdrawShares"] != 1 {
					t.Errorf("pendingWithdrawShares called %d times, want 1", mc.callCounts["pendingWithdrawShares"])
				}
			},
		},
		{
			name: "paused vault and non-empty withdraw queue: scan skips zero-share slots",
			mock: &mockCaller{
				totalAssets:           big.NewInt(42),
				idleAssets:            big.NewInt(42),
				idleReserveBps:        0,
				maxRebalanceDelta:     0,
				paused:                true,
				strategyIDs:           [][32]byte{}, // no strategies
				pendingWithdrawShares: big.NewInt(900),
				nextWithdrawRequestID: big.NewInt(4), // scan ids 0,1,2,3
				targetByID:            map[[32]byte]uint16{},
				capByID:               map[[32]byte]uint16{},
				assetsByID:            map[[32]byte]*big.Int{},
				quarantinedByID:       map[[32]byte]bool{},
				withdrawByID: map[string]vault.LibWithdrawQueueWithdrawRequest{
					// id 0: empty/fulfilled (shares 0) -> skipped
					"1": {Owner: common.HexToAddress("0x1"), Receiver: common.HexToAddress("0x2"), Shares: big.NewInt(600)},
					// id 2: cancelled (shares 0) -> skipped
					"3": {Owner: common.HexToAddress("0x3"), Receiver: common.HexToAddress("0x4"), Shares: big.NewInt(300)},
				},
			},
			check: func(t *testing.T, mc *mockCaller, st *types.VaultState) {
				if !st.Paused {
					t.Errorf("Paused = false, want true")
				}
				if len(st.Strategies) != 0 {
					t.Errorf("len(Strategies) = %d, want 0", len(st.Strategies))
				}
				// queue non-empty -> scan happened.
				if mc.callCounts["nextWithdrawRequestId"] != 1 {
					t.Errorf("nextWithdrawRequestId called %d times, want 1", mc.callCounts["nextWithdrawRequestId"])
				}
				if mc.callCounts["withdrawRequest"] != 4 {
					t.Errorf("withdrawRequest called %d times, want 4 (ids 0..3)", mc.callCounts["withdrawRequest"])
				}
				if len(st.PendingWithdraws) != 2 {
					t.Fatalf("len(PendingWithdraws) = %d, want 2 (zero-share slots skipped)", len(st.PendingWithdraws))
				}
				if st.PendingWithdraws[0].ID.Cmp(big.NewInt(1)) != 0 || st.PendingWithdraws[0].Shares.Cmp(big.NewInt(600)) != 0 {
					t.Errorf("PendingWithdraws[0] = {ID:%s Shares:%s}, want {1 600}", st.PendingWithdraws[0].ID, st.PendingWithdraws[0].Shares)
				}
				if st.PendingWithdraws[1].ID.Cmp(big.NewInt(3)) != 0 || st.PendingWithdraws[1].Shares.Cmp(big.NewInt(300)) != 0 {
					t.Errorf("PendingWithdraws[1] = {ID:%s Shares:%s}, want {3 300}", st.PendingWithdraws[1].ID, st.PendingWithdraws[1].Shares)
				}
			},
		},
	}

	parsedABI := mustABI(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock.abi = parsedABI
			tc.mock.callCounts = map[string]int{}

			r, err := NewChainReader(tc.mock, vaultAddr)
			if err != nil {
				t.Fatalf("NewChainReader: %v", err)
			}
			st, err := r.Snapshot(context.Background())
			if err != nil {
				t.Fatalf("Snapshot: %v", err)
			}
			tc.check(t, tc.mock, st)
		})
	}
}
