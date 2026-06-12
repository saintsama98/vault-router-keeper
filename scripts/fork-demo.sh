#!/usr/bin/env bash
# =============================================================================
# fork-demo.sh — realtime end-to-end demo on an Arbitrum One fork.
#
# Stands up: anvil fork -> real-facet diamond (DeployFork.s.sol in the diamond
# repo) -> keeper with composite (LIVE) risk layer -> simulated users
# depositing/withdrawing real USDC while the keeper allocates into real
# Aave/Pendle/Morpho positions.
#
# Prereqs:
#   - diamond repo has script/DeployFork.s.sol   (diamond-side deliverable)
#   - .env.fork in this repo has FILL-FROM-REPORT fields populated
#
# Usage:   ./scripts/fork-demo.sh [fork-rpc-url]
# Cleanup: the script kills anvil + keeper on exit (trap).
# =============================================================================
set -euo pipefail

FOUNDRY=/mnt/adiii_dev/dev_env/foundry/bin
DIAMOND_REPO=/mnt/adiii_dev/Ethereum-dev/vault-router-diamond
KEEPER_REPO="$(cd "$(dirname "$0")/.." && pwd)"

FORK_URL="${1:-https://arb1.arbitrum.io/rpc}"
PORT=8549
RPC="http://127.0.0.1:${PORT}"

USDC=0xaf88d065e77c8cC2239327C5EDb3A432268e5831
# aUSDC holds the Aave pool's USDC liquidity — impersonated as the demo's
# USDC faucet (fork-only trick; never touches a real key).
USDC_WHALE=0x724dc807b04555b71ed48a6896b6F41593b8C637

DEPLOYER_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
# anvil accounts #1..#3 — the fake users.
ALICE=0x70997970C51812dc3A010C7d01b50e0d17dc79C8
ALICE_KEY=0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
BOB=0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
BOB_KEY=0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
CAROL=0x90F79bf6EB2c4f870365E785982E1f101E93b906
CAROL_KEY=0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6

STRAT_IDS=(
  0x6161766500000000000000000000000000000000000000000000000000000000 # aave
  0x6d6f7270686f0000000000000000000000000000000000000000000000000000 # morpho
  0x70656e646c650000000000000000000000000000000000000000000000000000 # pendle
)
STRAT_NAMES=(aave morpho pendle)

cast() { "$FOUNDRY/cast" "$@"; }
log()  { printf '\n\033[1;36m== %s ==\033[0m\n' "$*"; }

PIDS=()
cleanup() { for p in "${PIDS[@]:-}"; do kill "$p" 2>/dev/null || true; done; }
trap cleanup EXIT

# --- 1. fork anvil -----------------------------------------------------------
log "starting anvil fork of Arbitrum One on :$PORT"
"$FOUNDRY/anvil" --fork-url "$FORK_URL" --port "$PORT" --chain-id 42161 --silent &
PIDS+=($!)
for _ in $(seq 1 30); do cast chain-id --rpc-url "$RPC" >/dev/null 2>&1 && break; sleep 1; done
cast chain-id --rpc-url "$RPC" >/dev/null || { echo "anvil not ready"; exit 1; }

# --- 2. deploy the real-facet diamond ---------------------------------------
log "deploying diamond via DeployFork.s.sol (Aave + Morpho; Pendle gated off — no USDC-compatible live market)"
[ -f "$DIAMOND_REPO/script/DeployFork.s.sol" ] || { echo "DeployFork.s.sol missing — diamond side not ready"; exit 1; }
# Gauntlet USDC Core (MetaMorpho, Arbitrum) — verified asset()==USDC, ~$3.2M TVL.
export ARB_MORPHO_VAULT=0x7e97fa6893871A2751B5fE961978DCCb2c201E65
(cd "$DIAMOND_REPO" && "$FOUNDRY/forge" script script/DeployFork.s.sol:DeployFork \
  --rpc-url "$RPC" --broadcast --private-key "$DEPLOYER_KEY" -vv)

VAULT=$(python3 - "$DIAMOND_REPO/broadcast/DeployFork.s.sol/42161/run-latest.json" <<'EOF'
import json, sys
txs = json.load(open(sys.argv[1]))["transactions"]
print(next(t["contractAddress"] for t in txs
           if t.get("transactionType") == "CREATE" and t.get("contractName") == "Vault"))
EOF
)
log "diamond (vault) at $VAULT"

# --- 3. fund the fake users with real USDC ----------------------------------
log "funding users with real USDC (impersonated aUSDC reserve)"
cast rpc anvil_impersonateAccount "$USDC_WHALE" --rpc-url "$RPC" >/dev/null
cast rpc anvil_setBalance "$USDC_WHALE" 0xDE0B6B3A7640000 --rpc-url "$RPC" >/dev/null
for u in "$ALICE:250000" "$BOB:100000" "$CAROL:50000"; do
  addr="${u%%:*}"; amt="${u##*:}"
  cast send "$USDC" "transfer(address,uint256)(bool)" "$addr" "${amt}000000" \
    --from "$USDC_WHALE" --unlocked --rpc-url "$RPC" >/dev/null
  echo "  $addr <- $amt USDC (bal: $(cast call "$USDC" "balanceOf(address)(uint256)" "$addr" --rpc-url "$RPC"))"
done
cast rpc anvil_stopImpersonatingAccount "$USDC_WHALE" --rpc-url "$RPC" >/dev/null

# --- 4. start the keeper (composite/live risk) -------------------------------
log "building + starting keeper (composite risk, live signing)"
(cd "$KEEPER_REPO" && GOWORK=off go build -o bin/keeper ./cmd/keeper)
set -a; . "$KEEPER_REPO/.env.fork"; set +a
export KEEPER_RPC_URL="$RPC" KEEPER_VAULT_ADDRESS="$VAULT"
"$KEEPER_REPO/bin/keeper" 2>&1 | sed 's/^/  [keeper] /' &
PIDS+=($!)

snapshot() {
  echo "  totalAssets: $(cast call "$VAULT" "totalAssets()(uint256)" --rpc-url "$RPC")  paused: $(cast call "$VAULT" "paused()(bool)" --rpc-url "$RPC")"
  for i in "${!STRAT_IDS[@]}"; do
    t=$(cast call "$VAULT" "targetAllocation(bytes32)(uint16)" "${STRAT_IDS[$i]}" --rpc-url "$RPC" 2>/dev/null || echo n/a)
    a=$(cast call "$VAULT" "strategyTotalAssets(bytes32)(uint256)" "${STRAT_IDS[$i]}" --rpc-url "$RPC" 2>/dev/null || echo n/a)
    echo "  ${STRAT_NAMES[$i]}: target=${t}bps assets=${a}"
  done
}

# --- 5. scenario: staggered deposits while the keeper allocates --------------
log "phase: initial state"; snapshot

deposit() { # user key amount(USDC)
  cast send "$USDC" "approve(address,uint256)(bool)" "$VAULT" "${3}000000" --private-key "$2" --rpc-url "$RPC" >/dev/null
  cast send "$VAULT" "deposit(uint256,address)(uint256)" "${3}000000" "$1" --private-key "$2" --rpc-url "$RPC" >/dev/null
  echo "  deposited $3 USDC for $1 (shares: $(cast call "$VAULT" "balanceOf(address)(uint256)" "$1" --rpc-url "$RPC"))"
}

log "phase: deposits (alice 250k, bob 100k, carol 50k — staggered)"
deposit "$ALICE" "$ALICE_KEY" 250000; sleep 12; snapshot
deposit "$BOB"   "$BOB_KEY"   100000; sleep 12; snapshot
deposit "$CAROL" "$CAROL_KEY" 50000;  sleep 20

log "phase: post-allocation state (keeper has had ~45s of ticks)"
snapshot

# --- 6. withdraw path --------------------------------------------------------
log "phase: carol withdraws half her shares (direct path from idle)"
# awk strips cast's "[5e13]" scientific-notation annotation so bash arithmetic works
CAROL_SHARES=$(cast call "$VAULT" "balanceOf(address)(uint256)" "$CAROL" --rpc-url "$RPC" | awk '{print $1}')
cast send "$VAULT" "redeem(uint256,address,address)(uint256)" $((CAROL_SHARES / 2)) "$CAROL" "$CAROL" \
  --private-key "$CAROL_KEY" --rpc-url "$RPC" >/dev/null \
  && echo "  redeem OK (carol USDC: $(cast call "$USDC" "balanceOf(address)(uint256)" "$CAROL" --rpc-url "$RPC"))" \
  || echo "  redeem REVERTED (share-lock or insufficient idle — check requestWithdraw path)"
snapshot

log "demo complete — keeper + anvil will be torn down"
