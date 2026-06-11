package execute

import (
	"context"
	"log/slog"
	"math/big"
	"sync"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/vault-router-keeper/internal/bindings/vault"
	"github.com/vault-router-keeper/pkg/types"
)

func id(b byte) types.StrategyID {
	var s types.StrategyID
	s[0] = b
	return s
}

// flattenAllocation must be deterministic: the same target map always yields the
// same id order (sorted) and matching bps — identical calldata every call
// (PLAN.md §6 invariant 4), independent of Go's random map iteration order.
func TestFlattenAllocationDeterministic(t *testing.T) {
	alloc := &types.Allocation{Targets: map[types.StrategyID]types.Bps{
		id(0x03): 1000,
		id(0x01): 5000,
		id(0x02): 0,
	}}

	wantIDs := []types.StrategyID{id(0x01), id(0x02), id(0x03)}
	wantBps := []uint16{5000, 0, 1000}

	for iter := range 50 {
		ids, bps := flattenAllocation(alloc)
		if len(ids) != 3 || len(bps) != 3 {
			t.Fatalf("len ids=%d bps=%d, want 3/3", len(ids), len(bps))
		}
		for i := range wantIDs {
			if ids[i] != [32]byte(wantIDs[i]) {
				t.Fatalf("iter %d pos %d: id=%x want %x", iter, i, ids[i][:4], wantIDs[i][:4])
			}
			if bps[i] != wantBps[i] {
				t.Fatalf("iter %d pos %d: bps=%d want %d", iter, i, bps[i], wantBps[i])
			}
		}
	}
}

func TestLoadKey(t *testing.T) {
	// Well-known test key (never used on any live network).
	const k = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"bare hex", k, true},
		{"0x prefixed", "0x" + k, true},
		{"empty", "", false},
		{"garbage", "nothex", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadKey(tc.in)
			if (err == nil) != tc.ok {
				t.Fatalf("LoadKey(%q) err=%v, want ok=%v", tc.in, err, tc.ok)
			}
		})
	}
}

// Well-known anvil/Hardhat test key (account #0). Never used on any live
// network; safe to embed. Matches the key the task pins for this test.
const testKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

// testChainID is anvil's default chain id.
var testChainID = big.NewInt(31337)

// vaultAddr is an arbitrary contract address the bound transactor targets.
var vaultAddr = common.HexToAddress("0x000000000000000000000000000000000000C0DE")

// mockBackend is the minimal bind.ContractBackend the LocalExecutor's transactor
// drives when it assembles + signs a tx. It returns canned fee/nonce/gas values
// so go-ethereum can build a valid EIP-1559 tx without a live node, and records
// every SendTransaction so the test can inspect the signed calldata.
//
// Method set is dictated by bind.ContractBackend in go-ethereum v1.17.3:
// ContractCaller{CodeAt,CallContract} + ContractTransactor{EstimateGas,
// SuggestGasPrice,SuggestGasTipCap,SendTransaction,HeaderByNumber,PendingCodeAt,
// PendingNonceAt} + ContractFilterer{FilterLogs,SubscribeFilterLogs}.
type mockBackend struct {
	mu   sync.Mutex
	sent []*gethtypes.Transaction // captured, in send order
}

func (m *mockBackend) lastSent() *gethtypes.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sent) == 0 {
		return nil
	}
	return m.sent[len(m.sent)-1]
}

// --- ContractTransactor ---

func (m *mockBackend) SendTransaction(_ context.Context, tx *gethtypes.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, tx)
	return nil
}

func (m *mockBackend) HeaderByNumber(_ context.Context, _ *big.Int) (*gethtypes.Header, error) {
	// BaseFee must be non-nil: createDynamicTx derives GasFeeCap from it.
	return &gethtypes.Header{Number: big.NewInt(1), BaseFee: big.NewInt(1_000_000_000)}, nil
}

func (m *mockBackend) PendingCodeAt(_ context.Context, _ common.Address) ([]byte, error) {
	// Non-empty so estimateGasLimit does not short-circuit with ErrNoCode.
	return []byte{0x60, 0x00}, nil
}

func (m *mockBackend) PendingNonceAt(_ context.Context, _ common.Address) (uint64, error) {
	return 7, nil
}

func (m *mockBackend) SuggestGasPrice(_ context.Context) (*big.Int, error) {
	return big.NewInt(1_000_000_000), nil
}

func (m *mockBackend) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	return big.NewInt(1_000_000_000), nil
}

func (m *mockBackend) EstimateGas(_ context.Context, _ ethereum.CallMsg) (uint64, error) {
	return 100_000, nil
}

// --- ContractCaller ---

func (m *mockBackend) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return []byte{0x60, 0x00}, nil
}

func (m *mockBackend) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return nil, nil
}

// --- ContractFilterer (unused by the transactor; present to satisfy the iface) ---

func (m *mockBackend) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]gethtypes.Log, error) {
	return nil, nil
}

func (m *mockBackend) SubscribeFilterLogs(_ context.Context, _ ethereum.FilterQuery, _ chan<- gethtypes.Log) (ethereum.Subscription, error) {
	return nil, nil
}

func newTestExecutor(t *testing.T, b Backend) *LocalExecutor {
	t.Helper()
	key, err := LoadKey(testKey)
	if err != nil {
		t.Fatalf("LoadKey: %v", err)
	}
	e, err := NewLocalExecutor(b, vaultAddr, key, testChainID, slog.Default())
	if err != nil {
		t.Fatalf("NewLocalExecutor: %v", err)
	}
	return e
}

// selector returns the 4-byte ABI selector for method against the real vault ABI,
// so the test asserts against the generated binding rather than a hard-coded value.
func selector(t *testing.T, method string) [4]byte {
	t.Helper()
	parsed, err := vault.VaultMetaData.GetAbi()
	if err != nil {
		t.Fatalf("parse vault abi: %v", err)
	}
	m, ok := parsed.Methods[method]
	if !ok {
		t.Fatalf("method %q not in vault abi", method)
	}
	var s [4]byte
	copy(s[:], m.ID)
	return s
}

func dataSelector(data []byte) [4]byte {
	var s [4]byte
	copy(s[:], data)
	return s
}

// Each non-guard ActionKind must send exactly one tx whose calldata selector
// matches the corresponding vault method on the generated binding.
func TestExecuteSelectorPerKind(t *testing.T) {
	cases := []struct {
		name   string
		action types.Action
		method string
	}{
		{
			"SetAllocation",
			types.Action{Kind: types.ActionSetAllocation, Allocation: &types.Allocation{
				Targets: map[types.StrategyID]types.Bps{id(0x02): 4000, id(0x01): 6000},
			}},
			"setAllocation",
		},
		{"Rebalance", types.Action{Kind: types.ActionRebalance}, "rebalance"},
		{"HarvestAll", types.Action{Kind: types.ActionHarvestAll}, "harvestAll"},
		{
			"FulfillWithdraw",
			types.Action{Kind: types.ActionFulfillWithdraw, WithdrawID: big.NewInt(42)},
			"fulfillWithdraw",
		},
		{"GuardCheckpoint", types.Action{Kind: types.ActionGuardCheckpoint}, "guardCheckpoint"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := &mockBackend{}
			e := newTestExecutor(t, b)

			hash, err := e.Execute(context.Background(), tc.action)
			if err != nil {
				t.Fatalf("Execute(%s): %v", tc.name, err)
			}
			if hash == "" {
				t.Fatalf("Execute(%s): empty hash", tc.name)
			}

			tx := b.lastSent()
			if tx == nil {
				t.Fatalf("Execute(%s): no transaction sent", tc.name)
			}
			if len(b.sent) != 1 {
				t.Fatalf("Execute(%s): sent %d txs, want 1", tc.name, len(b.sent))
			}
			// Returned hash must be the hash of the tx actually sent.
			if got := tx.Hash().Hex(); got != hash {
				t.Fatalf("Execute(%s): returned hash %s != sent tx hash %s", tc.name, hash, got)
			}
			// Tx must target the vault and be EIP-155 signed for the right chain.
			if to := tx.To(); to == nil || *to != vaultAddr {
				t.Fatalf("Execute(%s): tx.To=%v, want %v", tc.name, to, vaultAddr)
			}
			if tx.ChainId().Cmp(testChainID) != 0 {
				t.Fatalf("Execute(%s): chainId %v, want %v", tc.name, tx.ChainId(), testChainID)
			}

			got := dataSelector(tx.Data())
			want := selector(t, tc.method)
			if got != want {
				t.Fatalf("Execute(%s): calldata selector %x, want %x (%s)", tc.name, got, want, tc.method)
			}
		})
	}
}

// SetAllocation must pack the sorted (ids,bps) arrays — the calldata after the
// selector must ABI-decode back to the deterministic flattenAllocation output.
func TestExecuteSetAllocationCalldata(t *testing.T) {
	alloc := &types.Allocation{Targets: map[types.StrategyID]types.Bps{
		id(0x03): 1000,
		id(0x01): 5000,
		id(0x02): 0,
	}}
	wantIDs, wantBps := flattenAllocation(alloc) // sorted, deterministic

	b := &mockBackend{}
	e := newTestExecutor(t, b)

	if _, err := e.Execute(context.Background(), types.Action{
		Kind:       types.ActionSetAllocation,
		Allocation: alloc,
	}); err != nil {
		t.Fatalf("Execute(SetAllocation): %v", err)
	}

	tx := b.lastSent()
	if tx == nil {
		t.Fatal("no tx sent")
	}
	if got, want := dataSelector(tx.Data()), selector(t, "setAllocation"); got != want {
		t.Fatalf("selector %x, want %x", got, want)
	}

	parsed, err := vault.VaultMetaData.GetAbi()
	if err != nil {
		t.Fatalf("parse abi: %v", err)
	}
	args, err := parsed.Methods["setAllocation"].Inputs.Unpack(tx.Data()[4:])
	if err != nil {
		t.Fatalf("unpack setAllocation args: %v", err)
	}
	if len(args) != 2 {
		t.Fatalf("got %d args, want 2", len(args))
	}
	gotIDs, ok := args[0].([][32]byte)
	if !ok {
		t.Fatalf("arg0 type %T, want [][32]byte", args[0])
	}
	gotBps, ok := args[1].([]uint16)
	if !ok {
		t.Fatalf("arg1 type %T, want []uint16", args[1])
	}
	if len(gotIDs) != len(wantIDs) || len(gotBps) != len(wantBps) {
		t.Fatalf("lens ids=%d/%d bps=%d/%d", len(gotIDs), len(wantIDs), len(gotBps), len(wantBps))
	}
	for i := range wantIDs {
		if gotIDs[i] != [32]byte(wantIDs[i]) {
			t.Fatalf("ids[%d]=%x want %x", i, gotIDs[i][:4], wantIDs[i][:4])
		}
		if gotBps[i] != wantBps[i] {
			t.Fatalf("bps[%d]=%d want %d", i, gotBps[i], wantBps[i])
		}
	}
}

// FulfillWithdraw must encode the *big.Int withdraw id into the calldata.
func TestExecuteFulfillWithdrawCalldata(t *testing.T) {
	const wantID = 0xABCDEF
	b := &mockBackend{}
	e := newTestExecutor(t, b)

	if _, err := e.Execute(context.Background(), types.Action{
		Kind:       types.ActionFulfillWithdraw,
		WithdrawID: big.NewInt(wantID),
	}); err != nil {
		t.Fatalf("Execute(FulfillWithdraw): %v", err)
	}

	tx := b.lastSent()
	if tx == nil {
		t.Fatal("no tx sent")
	}
	parsed, err := vault.VaultMetaData.GetAbi()
	if err != nil {
		t.Fatalf("parse abi: %v", err)
	}
	args, err := parsed.Methods["fulfillWithdraw"].Inputs.Unpack(tx.Data()[4:])
	if err != nil {
		t.Fatalf("unpack fulfillWithdraw args: %v", err)
	}
	if len(args) != 1 {
		t.Fatalf("got %d args, want 1", len(args))
	}
	gotID, ok := args[0].(*big.Int)
	if !ok {
		t.Fatalf("arg0 type %T, want *big.Int", args[0])
	}
	if gotID.Int64() != wantID {
		t.Fatalf("withdraw id %d, want %d", gotID.Int64(), wantID)
	}
}

// The nil-allocation and nil-withdraw-id guards must error WITHOUT sending a tx,
// and an unknown kind must also error without sending.
func TestExecuteGuardsDoNotSend(t *testing.T) {
	cases := []struct {
		name   string
		action types.Action
	}{
		{"nil allocation", types.Action{Kind: types.ActionSetAllocation, Allocation: nil}},
		{"nil withdraw id", types.Action{Kind: types.ActionFulfillWithdraw, WithdrawID: nil}},
		{"unknown kind", types.Action{Kind: types.ActionKind(99)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := &mockBackend{}
			e := newTestExecutor(t, b)

			hash, err := e.Execute(context.Background(), tc.action)
			if err == nil {
				t.Fatalf("Execute(%s): want error, got nil (hash=%q)", tc.name, hash)
			}
			if hash != "" {
				t.Fatalf("Execute(%s): want empty hash on guard, got %q", tc.name, hash)
			}
			if len(b.sent) != 0 {
				t.Fatalf("Execute(%s): guard sent %d txs, want 0", tc.name, len(b.sent))
			}
		})
	}
}
