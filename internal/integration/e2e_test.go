//go:build integration

// Package integration holds the self-contained, fresh-anvil end-to-end test that
// exercises the WHOLE keeper against a REAL diamond: it stands up a fresh anvil,
// deploys the Vault Router diamond via the diamond repo's forge script, then
// builds and runs the keeper binary for ~one tick with a live signing executor,
// and asserts the on-chain targetAllocation CHANGED — proving the full
// perceive -> decide -> validate -> sign -> send loop end to end.
//
// It is gated behind the `integration` build tag so the normal gate
// (go build/vet/test ./...) never compiles or runs it, and it t.Skip()s cleanly
// when anvil/forge/the diamond repo are unavailable so CI can never falsely fail.
//
// Run:
//
//	GOWORK=off go test -tags integration ./internal/integration/... -run E2E -v
//
// See the RUNBOOK comment at the bottom for the --fork-url variant.
package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/vault-router-keeper/internal/perceive"
	"github.com/vault-router-keeper/pkg/types"
)

const (
	// foundryBin is where anvil/forge/cast live in this environment.
	foundryBin = "/mnt/adiii_dev/dev_env/foundry/bin"
	// diamondRepo is the Solidity repo whose script/Deploy.s.sol assembles the
	// diamond against the repo's own mocks.
	diamondRepo = "/mnt/adiii_dev/Ethereum-dev/vault-router-diamond"

	// anvilPort is a non-default port so the test never collides with a dev anvil
	// on 8545.
	anvilPort = "8547"
	rpcURL    = "http://localhost:" + anvilPort
	chainID   = "31337"

	// anvilAcct0Key is the well-known fresh-anvil account #0 key. It is the diamond
	// owner AND (by Deploy.s.sol's CURATOR default) the curator, so the keeper
	// signing with it is curator out of the box. This is a throwaway local test key,
	// never a real secret.
	anvilAcct0Key = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	// mockStrategyID is bytes32("mock") — the single strategy Deploy.s.sol registers
	// (logged via console2.logBytes32). The keeper allocates into it.
	mockStrategyID = "0x6d6f636b00000000000000000000000000000000000000000000000000000000"
)

// TestE2EFreshAnvilRebalance is the full read->decide->validate->sign->send proof.
func TestE2EFreshAnvilRebalance(t *testing.T) {
	anvil := filepath.Join(foundryBin, "anvil")
	forge := filepath.Join(foundryBin, "forge")
	cast := filepath.Join(foundryBin, "cast")

	// --- preflight: skip cleanly when the toolchain/repo is unavailable --------
	for _, bin := range []string{anvil, forge, cast} {
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("foundry tool missing (%s): %v", bin, err)
		}
	}
	if _, err := os.Stat(filepath.Join(diamondRepo, "script", "Deploy.s.sol")); err != nil {
		t.Skipf("diamond repo / Deploy.s.sol unavailable: %v", err)
	}

	// --- (a) start a FRESH anvil in the background -----------------------------
	anvilCtx, stopAnvil := context.WithCancel(context.Background())
	anvilCmd := exec.CommandContext(anvilCtx, anvil,
		"--port", anvilPort, "--chain-id", chainID, "--silent")
	// Cancel SIGKILLs anvil and WaitDelay bounds the Wait() so a stuck child can
	// never wedge teardown (the cause of an earlier 10-min test-timeout hang).
	anvilCmd.Cancel = func() error { return anvilCmd.Process.Kill() }
	anvilCmd.WaitDelay = 5 * time.Second
	if err := anvilCmd.Start(); err != nil {
		t.Skipf("cannot start anvil: %v", err)
	}
	// (f) tear down anvil: cancel FIRST (kill), THEN reap. A single defer keeps the
	// kill-before-wait ordering; two separate defers would Wait() while the process
	// is still alive (LIFO) and block forever.
	defer func() {
		stopAnvil()
		_ = anvilCmd.Wait()
	}()

	if err := waitForAnvil(t, cast); err != nil {
		t.Skipf("anvil did not become ready: %v", err)
	}

	// --- (b) deploy the diamond via the forge script ---------------------------
	deployCtx, cancelDeploy := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelDeploy()
	deploy := exec.CommandContext(deployCtx, forge, "script",
		"script/Deploy.s.sol:Deploy",
		"--rpc-url", rpcURL,
		"--broadcast",
		"--private-key", anvilAcct0Key,
	)
	deploy.Dir = diamondRepo
	deployOut, err := deploy.CombinedOutput()
	if err != nil {
		t.Fatalf("forge script deploy failed: %v\n%s", err, truncate(deployOut, 4000))
	}
	t.Logf("forge script deploy OK")

	// --- (c) parse the deployed diamond address --------------------------------
	// Prefer the broadcast file (machine-readable); fall back to the console log.
	vaultAddr := vaultFromBroadcast(t)
	if vaultAddr == "" {
		vaultAddr = vaultFromConsole(deployOut)
	}
	if vaultAddr == "" {
		t.Fatalf("could not discover diamond address from broadcast or console:\n%s", truncate(deployOut, 2000))
	}
	t.Logf("diamond (vault) at %s", vaultAddr)

	// --- read the INITIAL on-chain target (must be 0 on a fresh deploy) --------
	before := targetAllocation(t, cast, vaultAddr)
	t.Logf("targetAllocation BEFORE keeper: %d bps", before)

	// --- (d) build + run the keeper for ~one tick ------------------------------
	keeperBin := buildKeeper(t)

	keeperCtx, cancelKeeper := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancelKeeper()
	keeper := exec.CommandContext(keeperCtx, keeperBin)
	keeper.Env = append(os.Environ(),
		"KEEPER_RPC_URL="+rpcURL,
		"KEEPER_VAULT_ADDRESS="+vaultAddr,
		"KEEPER_CHAIN_ID="+chainID,
		"KEEPER_DRY_RUN=false",
		"KEEPER_BRAIN=gauntlet-lite",
		"KEEPER_RISK_PROVIDER=static",
		"KEEPER_PRIVATE_KEY="+anvilAcct0Key,
		// Short cadences so the first tick fires guard+rebalance+harvest immediately.
		"KEEPER_POLL_INTERVAL=2s",
		"KEEPER_REBALANCE_MIN_INTERVAL=1s",
		"KEEPER_GUARD_INTERVAL=1s",
		"KEEPER_HARVEST_INTERVAL=1s",
	)
	keeperOut, _ := keeper.CombinedOutput() // killed by ctx timeout => non-nil err is expected
	t.Logf("keeper output (tail):\n%s", tailLines(string(keeperOut), 24))

	// --- (e) assert the on-chain target CHANGED --------------------------------
	// Cross-check two independent read paths: a fresh perceive.ChainReader
	// Snapshot (the keeper's own read code) AND a raw cast call.
	afterCast := targetAllocation(t, cast, vaultAddr)
	afterReader := targetViaChainReader(t, vaultAddr)
	t.Logf("targetAllocation AFTER keeper: cast=%d bps, ChainReader=%d bps", afterCast, afterReader)

	if afterCast != afterReader {
		t.Fatalf("read paths disagree: cast=%d != ChainReader=%d", afterCast, afterReader)
	}
	if afterCast == before {
		t.Fatalf("targetAllocation did NOT change (before=%d after=%d): "+
			"the read->decide->validate->sign->send loop did not land a SetAllocation.\n"+
			"keeper output:\n%s", before, afterCast, tailLines(string(keeperOut), 40))
	}
	// The single mock strategy has a 60%% cap and a 20%% idle floor, so the
	// gauntlet-lite water-fill targets the 6000 bps cap. We assert "changed and
	// moved up", not an exact magic number, to stay robust to model tuning.
	if afterCast <= before {
		t.Fatalf("expected target to increase from %d, got %d", before, afterCast)
	}
	t.Logf("PASS: targetAllocation %d -> %d bps (read->decide->validate->sign->send proven E2E)", before, afterCast)
}

// waitForAnvil polls `cast block-number` until the node answers or times out.
func waitForAnvil(t *testing.T, cast string) error {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		out, err := exec.CommandContext(ctx, cast, "block-number", "--rpc-url", rpcURL).CombinedOutput()
		cancel()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for anvil at %s", rpcURL)
}

// vaultFromBroadcast reads the standard forge broadcast file and returns the
// contractAddress of the CREATE tx whose contractName is "Vault". Empty on any
// miss so the caller can fall back to the console log.
func vaultFromBroadcast(t *testing.T) string {
	t.Helper()
	path := filepath.Join(diamondRepo, "broadcast", "Deploy.s.sol", chainID, "run-latest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Logf("broadcast file unavailable (%s): %v; falling back to console", path, err)
		return ""
	}
	var doc struct {
		Transactions []struct {
			TransactionType string `json:"transactionType"`
			ContractName    string `json:"contractName"`
			ContractAddress string `json:"contractAddress"`
		} `json:"transactions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Logf("broadcast file parse error: %v; falling back to console", err)
		return ""
	}
	for _, tx := range doc.Transactions {
		if tx.ContractName == "Vault" && strings.EqualFold(tx.TransactionType, "CREATE") && tx.ContractAddress != "" {
			return common.HexToAddress(tx.ContractAddress).Hex()
		}
	}
	return ""
}

// vaultFromConsole scrapes "Diamond (Vault) deployed at: <addr>" from the forge
// script stdout as a fallback discovery path.
func vaultFromConsole(out []byte) string {
	const marker = "Diamond (Vault) deployed at:"
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, marker); i >= 0 {
			rest := strings.TrimSpace(line[i+len(marker):])
			for _, f := range strings.Fields(rest) {
				if strings.HasPrefix(f, "0x") && len(f) >= 42 {
					return common.HexToAddress(f).Hex()
				}
			}
		}
	}
	return ""
}

// targetAllocation reads the on-chain target via `cast call` (the independent,
// non-Go read path).
func targetAllocation(t *testing.T, cast, vaultAddr string) uint64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, cast, "call", vaultAddr,
		"targetAllocation(bytes32)(uint16)", mockStrategyID, "--rpc-url", rpcURL).CombinedOutput()
	if err != nil {
		t.Fatalf("cast targetAllocation failed: %v\n%s", err, out)
	}
	// cast may render uint16 as a decimal optionally followed by a bracketed hint.
	field := strings.Fields(strings.TrimSpace(string(out)))
	if len(field) == 0 {
		t.Fatalf("empty cast output for targetAllocation")
	}
	v, err := parseUint(field[0])
	if err != nil {
		t.Fatalf("parse cast targetAllocation %q: %v", string(out), err)
	}
	return v
}

// targetViaChainReader reads the same target through the keeper's own
// perceive.ChainReader.Snapshot, proving the production read path agrees.
func targetViaChainReader(t *testing.T, vaultAddr string) uint64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cl, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		t.Fatalf("dial for ChainReader: %v", err)
	}
	defer cl.Close()
	r, err := perceive.NewChainReader(cl, common.HexToAddress(vaultAddr))
	if err != nil {
		t.Fatalf("build ChainReader: %v", err)
	}
	state, err := r.Snapshot(ctx)
	if err != nil {
		t.Fatalf("ChainReader.Snapshot: %v", err)
	}
	want := types.StrategyID(common.HexToHash(mockStrategyID))
	for _, s := range state.Strategies {
		if s.ID == want {
			return uint64(s.TargetBps)
		}
	}
	t.Fatalf("mock strategy %s not found in snapshot (%d strategies)", mockStrategyID, len(state.Strategies))
	return 0
}

// buildKeeper compiles the keeper binary into a temp dir (GOWORK=off, matching the
// gate) and returns its path.
func buildKeeper(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	bin := filepath.Join(t.TempDir(), "keeper")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/keeper")
	build.Dir = repoRoot
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build keeper failed: %v\n%s", err, out)
	}
	return bin
}

// --- small helpers ----------------------------------------------------------

func parseUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	var v uint64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "\n...[truncated]"
}

func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// _ keeps math/big imported for callers that extend the assertion to *big.Int
// reads (e.g. CurrentAssets / TotalAssets) without re-touching the import block.
var _ = big.NewInt

// =============================================================================
// RUNBOOK — fresh-anvil vs --fork-url
// =============================================================================
//
// This test runs the FRESH-ANVIL variant: a clean local chain with the diamond
// assembled against the repo's own mocks (MockUSDC + MockProtocol +
// MockStrategyFacet). It is hermetic, fast, and needs no network/RPC secrets.
//
//	GOWORK=off go test -tags integration ./internal/integration/... -run E2E -v
//	# or:
//	make test-integration
//
// Manual equivalent (handy for debugging a single stage):
//
//	# 1. fresh chain
//	/mnt/adiii_dev/dev_env/foundry/bin/anvil --port 8547 --chain-id 31337
//	# 2. deploy
//	cd /mnt/adiii_dev/Ethereum-dev/vault-router-diamond
//	forge script script/Deploy.s.sol:Deploy --rpc-url http://localhost:8547 \
//	  --broadcast --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
//	# 3. run keeper one tick (KEEPER_DRY_RUN=false, gauntlet-lite, static risk)
//	# 4. assert: cast call <vault> "targetAllocation(bytes32)(uint16)" 0x6d6f636b...
//
// FORK-URL variant (testing against real Arbitrum protocol state — Aave V3 supply
// APR, Pendle oracles, live USDC) instead of mocks:
//
//	anvil --fork-url $ARBITRUM_RPC_URL --chain-id 42161 --port 8547
//
// then deploy/point the keeper with KEEPER_RISK_PROVIDER=composite and the live
// adapter env (KEEPER_AAVE_DATA_PROVIDER / KEEPER_AAVE_ASSET / KEEPER_AAVE_STRATEGY_IDS
// / KEEPER_PENDLE_ORACLE / KEEPER_PENDLE_MARKETS). The fork variant exercises the
// live risk/APY overlays; the fresh-anvil variant above exercises the control loop
// deterministically. Note: the mock-based Deploy.s.sol assembles its OWN asset and
// strategy, so on a fork it still uses mocks — the fork only matters once the
// strategy facets point at real protocol addresses (future work).
