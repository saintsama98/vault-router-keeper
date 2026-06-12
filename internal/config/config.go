// Package config loads keeper settings from environment variables. Secrets
// (the curator private key) are referenced by env-var NAME only — never stored
// here — so the key is read at runtime from a separate variable.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vault-router-keeper/pkg/types"
)

type Config struct {
	RPCURL       string
	VaultAddress string
	ChainID      int64
	KeyEnv       string // name of the env var holding the curator private key
	DryRun       bool
	Brain        string // which Decider: "stub" (default) | "gauntlet-lite"

	PollInterval         time.Duration
	HarvestInterval      time.Duration
	GuardInterval        time.Duration
	RebalanceMinInterval time.Duration

	// RiskProvider selects the EL signal layer wired into the gauntlet-lite brain:
	//   "static"    (default) — empty StaticRiskProvider (current behaviour; closed-form floor only).
	//   "composite" — route each StrategyID to its live adapter (Pendle/Aave/Credora).
	// Every live adapter degrades to ok=false when its external address/endpoint is
	// unconfigured, so "composite" with no addresses behaves exactly like "static".
	RiskProvider string

	// --- Aave V3 live reserve-risk reader (AaveChaosReader) ---------------------
	// AaveDataProvider is the Aave V3 ProtocolDataProvider address (allowed
	// hardcoded-external-address operator exception). Empty => Aave adapter ok=false.
	AaveDataProvider string
	// AaveAsset is the reserve to query == Vault.asset() (USDC on Arbitrum). The
	// curated vault ABI carries no asset() getter, so the operator supplies it.
	AaveAsset string
	// AaveStrategyIDs are the bytes32 StrategyIDs (hex) the Aave adapter owns.
	AaveStrategyIDs []types.StrategyID

	// --- Pendle PT live oracle/risk reader -------------------------------------
	// PendleOracle is the shared PendlePYLpOracle address. Empty => Pendle ok=false.
	PendleOracle string
	// PendleMarkets maps a StrategyID to its on-chain handles (market:pt:duration).
	PendleMarkets map[types.StrategyID]PendleMarket

	// --- Credora (Morpho facet) GraphQL risk feed ------------------------------
	// CredoraEndpoint is the operator-supplied Credora GraphQL URL. Empty =>
	// Credora ok=false (the public Morpho API exposes NO rating field — see
	// internal/risk/credora; do NOT point this at api.morpho.org/graphql).
	CredoraEndpoint string
	// CredoraAPIKey is sent as a Bearer token (the real Credora API is auth-gated).
	CredoraAPIKey string
	// CredoraMarkets maps a StrategyID to its Morpho market key.
	CredoraMarkets map[types.StrategyID]string

	// --- Morpho (MetaMorpho) public Blue-API yield feed ------------------------
	// MorphoAPIEndpoint is Morpho's public GraphQL URL (no auth — unlike Credora,
	// which stays the Morpho RISK source). Empty => yield adapter dormant and the
	// brain sees Morpho at APY 0 (residual-budget-only in the water-fill).
	MorphoAPIEndpoint string
	// MorphoVaults maps a StrategyID to its MetaMorpho ERC-4626 vault address.
	MorphoVaults map[types.StrategyID]string
}

// PendleMarket is the per-strategy Pendle handle parsed from KEEPER_PENDLE_MARKETS.
type PendleMarket struct {
	Market   string // Pendle PT/SY market address (queried via the oracle)
	PT       string // PrincipalToken address (queried for expiry)
	Duration uint32 // TWAP window, seconds
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() *Config {
	return &Config{
		RPCURL:               env("KEEPER_RPC_URL", ""),
		VaultAddress:         env("KEEPER_VAULT_ADDRESS", ""),
		ChainID:              envInt("KEEPER_CHAIN_ID", 42161), // Arbitrum One
		KeyEnv:               env("KEEPER_KEY_ENV", "KEEPER_PRIVATE_KEY"),
		DryRun:               envBool("KEEPER_DRY_RUN", true),
		Brain:                env("KEEPER_BRAIN", "stub"),
		PollInterval:         envDur("KEEPER_POLL_INTERVAL", 30*time.Second),
		HarvestInterval:      envDur("KEEPER_HARVEST_INTERVAL", 6*time.Hour),
		GuardInterval:        envDur("KEEPER_GUARD_INTERVAL", 5*time.Minute),
		RebalanceMinInterval: envDur("KEEPER_REBALANCE_MIN_INTERVAL", time.Hour),

		RiskProvider: env("KEEPER_RISK_PROVIDER", "static"),

		AaveDataProvider: env("KEEPER_AAVE_DATA_PROVIDER", ""),
		AaveAsset:        env("KEEPER_AAVE_ASSET", ""),
		AaveStrategyIDs:  parseStrategyIDs(env("KEEPER_AAVE_STRATEGY_IDS", "")),

		PendleOracle:  env("KEEPER_PENDLE_ORACLE", ""),
		PendleMarkets: parsePendleMarkets(env("KEEPER_PENDLE_MARKETS", "")),

		CredoraEndpoint: env("KEEPER_CREDORA_ENDPOINT", ""),
		CredoraAPIKey:   env(env("KEEPER_CREDORA_API_KEY_ENV", "KEEPER_CREDORA_API_KEY"), ""),
		CredoraMarkets:  parseStringMap(env("KEEPER_CREDORA_MARKETS", "")),

		MorphoAPIEndpoint: env("KEEPER_MORPHO_API", ""),
		MorphoVaults:      parseStringMap(env("KEEPER_MORPHO_VAULTS", "")),
	}
}

// parseStrategyIDs parses a comma-separated list of 32-byte hex StrategyIDs
// (0x-prefixed or bare). Malformed entries are skipped. Empty input => nil.
func parseStrategyIDs(s string) []types.StrategyID {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []types.StrategyID
	for _, tok := range strings.Split(s, ",") {
		if id, ok := parseStrategyID(strings.TrimSpace(tok)); ok {
			out = append(out, id)
		}
	}
	return out
}

// parseStrategyID decodes one 32-byte hex string into a StrategyID. Returns
// ok=false on any malformed/short input (so a typo degrades to no-route, not a
// panic or a fabricated id).
func parseStrategyID(s string) (types.StrategyID, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	s = strings.TrimPrefix(s, "0X")
	if len(s) != 64 {
		return types.StrategyID{}, false
	}
	var id types.StrategyID
	for i := 0; i < 32; i++ {
		b, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return types.StrategyID{}, false
		}
		id[i] = byte(b)
	}
	return id, true
}

// parseStringMap parses a CSV of "id=value" pairs into a StrategyID→string map,
// where id is a 32-byte hex StrategyID. Used for the Credora market-key mapping.
// Empty input => nil; malformed entries are skipped.
func parseStringMap(s string) map[types.StrategyID]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := make(map[types.StrategyID]string)
	for _, pair := range strings.Split(s, ",") {
		k, v, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			continue
		}
		if id, ok := parseStrategyID(strings.TrimSpace(k)); ok {
			out[id] = strings.TrimSpace(v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parsePendleMarkets parses a CSV of "id=market:pt:duration" entries into a
// StrategyID→PendleMarket map. market/pt are addresses, duration is seconds.
// Empty input => nil; malformed entries are skipped.
func parsePendleMarkets(s string) map[types.StrategyID]PendleMarket {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := make(map[types.StrategyID]PendleMarket)
	for _, pair := range strings.Split(s, ",") {
		k, v, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			continue
		}
		id, ok := parseStrategyID(strings.TrimSpace(k))
		if !ok {
			continue
		}
		parts := strings.Split(strings.TrimSpace(v), ":")
		if len(parts) != 3 {
			continue
		}
		dur, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 32)
		if err != nil {
			continue
		}
		out[id] = PendleMarket{
			Market:   strings.TrimSpace(parts[0]),
			PT:       strings.TrimSpace(parts[1]),
			Duration: uint32(dur),
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}

func envInt(k string, def int64) int64 {
	if v, ok := os.LookupEnv(k); ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v, ok := os.LookupEnv(k); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func envDur(k string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(k); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
