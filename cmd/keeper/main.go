// Command keeper runs the Vault Router keeper loop.
package main

import (
	"context"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/vault-router-keeper/internal/brain"
	"github.com/vault-router-keeper/internal/chain"
	"github.com/vault-router-keeper/internal/config"
	"github.com/vault-router-keeper/internal/execute"
	"github.com/vault-router-keeper/internal/keeper"
	"github.com/vault-router-keeper/internal/perceive"
	"github.com/vault-router-keeper/internal/risk"
	"github.com/vault-router-keeper/internal/risk/aave"
	"github.com/vault-router-keeper/internal/risk/credora"
	"github.com/vault-router-keeper/internal/risk/morpho"
	"github.com/vault-router-keeper/internal/risk/pendle"
	"github.com/vault-router-keeper/internal/trigger"
	"github.com/vault-router-keeper/internal/validate"
	"github.com/vault-router-keeper/pkg/types"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	log.Info("config",
		"rpc", cfg.RPCURL,
		"vault", cfg.VaultAddress,
		"chain", cfg.ChainID,
		"dryRun", cfg.DryRun,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// One RPC handle (when configured) backs both the live reader and executor.
	// *ethclient.Client satisfies bind.ContractCaller AND bind.ContractBackend.
	var client *ethclient.Client
	if cfg.RPCURL != "" {
		cl, err := chain.Dial(ctx, cfg.RPCURL)
		if err != nil {
			log.Error("dial RPC", "err", err)
			os.Exit(1)
		}
		defer cl.Close()
		client = cl
	}

	reader := buildReader(cfg, client, log)

	var decider brain.Decider
	switch cfg.Brain {
	case "gauntlet-lite":
		rp := buildRiskProvider(cfg, client, log)
		// Overlay live APY onto the snapshot (only the gauntlet-lite brain ranks
		// by yield; the stub echoes targets). Pass-through when no live yield feed.
		if yp := buildYieldProvider(cfg, client, log); yp != nil {
			reader = perceive.NewOverlayReader(reader, yp)
		}
		decider = brain.NewGauntletLiteDecider(rp, brain.DefaultRiskModel(), log)
	default:
		decider = brain.NewStubDecider()
	}
	sched := trigger.NewIntervalScheduler(map[string]time.Duration{
		"guard":     cfg.GuardInterval,
		"harvest":   cfg.HarvestInterval,
		"rebalance": cfg.RebalanceMinInterval,
	})

	validator := validate.New(log)
	exec := buildExecutor(cfg, client, log)

	k := keeper.New(reader, decider, validator, sched, exec, cfg.PollInterval, log)

	if err := k.Run(ctx); err != nil && ctx.Err() == nil {
		log.Error("keeper exited", "err", err)
		os.Exit(1)
	}
}

// buildReader returns the live ChainReader when an RPC client and vault address
// are configured, otherwise the empty StubReader (offline/dev loop).
func buildReader(cfg *config.Config, client *ethclient.Client, log *slog.Logger) perceive.Reader {
	if client == nil || cfg.VaultAddress == "" {
		log.Warn("no RPC/vault configured; using StubReader (empty state)")
		return perceive.NewStubReader()
	}
	r, err := perceive.NewChainReader(client, common.HexToAddress(cfg.VaultAddress))
	if err != nil {
		log.Error("build ChainReader", "err", err)
		os.Exit(1)
	}
	log.Info("perceive: live ChainReader", "vault", cfg.VaultAddress)
	return r
}

// buildRiskProvider builds the EL signal layer wired into the gauntlet-lite
// brain. The default ("static") preserves the prior behaviour: an empty
// StaticRiskProvider => closed-form EL floor for every strategy. With
// KEEPER_RISK_PROVIDER=composite, each StrategyID is routed to its LIVE adapter
// (Pendle on-chain / Aave ProtocolDataProvider / Credora GraphQL). Every adapter
// degrades to ok=false when its external address/endpoint is unconfigured, so
// "composite" with no addresses is observably identical to "static" — the honest
// no-data path, never a stub.
func buildRiskProvider(cfg *config.Config, client *ethclient.Client, log *slog.Logger) brain.RiskProvider {
	if cfg.RiskProvider != "composite" {
		return brain.NewStaticRiskProvider(nil)
	}

	routes := make(map[types.StrategyID]brain.RiskProvider)

	// Pendle: live maturity-structural liquidity gate over the on-chain oracle.
	if client != nil && cfg.PendleOracle != "" && len(cfg.PendleMarkets) > 0 {
		onchain := pendle.NewOnChain(client, common.HexToAddress(cfg.PendleOracle), pendleMarketConfigs(cfg))
		ids := make([]types.StrategyID, 0, len(cfg.PendleMarkets))
		for id := range cfg.PendleMarkets {
			ids = append(ids, id)
		}
		pr := pendle.NewReader(onchain, pendle.DefaultPolicy(), ids, nil)
		for _, id := range ids {
			routes[id] = pr
		}
		log.Info("risk: pendle live reader wired", "strategies", len(ids))
	}

	// Aave: live reserve-risk reader over the ProtocolDataProvider. Degrades to
	// ok=false when client/provider absent.
	if len(cfg.AaveStrategyIDs) > 0 {
		isAave := make(map[types.StrategyID]bool, len(cfg.AaveStrategyIDs))
		for _, id := range cfg.AaveStrategyIDs {
			isAave[id] = true
		}
		ar := aave.NewAaveChaosReader(callerOrNil(client), common.HexToAddress(cfg.AaveDataProvider), common.HexToAddress(cfg.AaveAsset), isAave)
		for _, id := range cfg.AaveStrategyIDs {
			routes[id] = ar
		}
		live := client != nil && cfg.AaveDataProvider != "" && cfg.AaveAsset != ""
		log.Info("risk: aave reader wired", "strategies", len(cfg.AaveStrategyIDs), "live", live)
	}

	// Credora (Morpho facet): live GraphQL rating feed. Degrades to ok=false when
	// the endpoint is empty (the public Morpho API has NO rating field).
	if len(cfg.CredoraMarkets) > 0 {
		cr := credora.NewReader(cfg.CredoraEndpoint, cfg.CredoraAPIKey, http.DefaultClient, cfg.CredoraMarkets)
		for id := range cfg.CredoraMarkets {
			routes[id] = cr
		}
		log.Info("risk: credora reader wired", "strategies", len(cfg.CredoraMarkets), "live", cfg.CredoraEndpoint != "")
	}

	if len(routes) == 0 {
		log.Warn("risk: composite selected but no adapters configured; closed-form floor only")
	}
	return risk.NewCompositeProvider(routes)
}

// buildYieldProvider builds the live APY overlay: the Pendle implied (fixed-to-
// maturity) APY plus the Aave V3 supply APR read from the ProtocolDataProvider
// (LiquidityRate). Each source is routed only for the StrategyIDs it owns and
// degrades to ok=false when its address/config is absent, so an empty config
// adds no route. Returns nil when no live yield source is configured at all, so
// the overlay is skipped entirely (APY stays 0 → risk-minimizing water-fill).
// Never fabricates a yield: every number traces to a live read.
func buildYieldProvider(cfg *config.Config, client *ethclient.Client, log *slog.Logger) perceive.YieldProvider {
	if cfg.RiskProvider != "composite" || client == nil {
		return nil
	}

	routes := make(map[types.StrategyID]perceive.YieldProvider)

	// Pendle: fixed-to-maturity implied APY over the on-chain oracle.
	if cfg.PendleOracle != "" && len(cfg.PendleMarkets) > 0 {
		onchain := pendle.NewOnChain(client, common.HexToAddress(cfg.PendleOracle), pendleMarketConfigs(cfg))
		ids := make([]types.StrategyID, 0, len(cfg.PendleMarkets))
		for id := range cfg.PendleMarkets {
			ids = append(ids, id)
		}
		yr := pendle.NewYieldReader(onchain, ids, nil)
		for _, id := range ids {
			routes[id] = yr
		}
		log.Info("perceive: pendle implied-APY overlay wired", "strategies", len(ids))
	}

	// Aave: live supply APR (LiquidityRate/1e27) over the same ProtocolDataProvider
	// address + asset + strategy ids used by the Aave risk reader. Degrades to
	// ok=false when client/provider/asset absent (empty config => no route).
	if len(cfg.AaveStrategyIDs) > 0 {
		isAave := make(map[types.StrategyID]bool, len(cfg.AaveStrategyIDs))
		for _, id := range cfg.AaveStrategyIDs {
			isAave[id] = true
		}
		yr := aave.NewYieldReader(callerOrNil(client), common.HexToAddress(cfg.AaveDataProvider), common.HexToAddress(cfg.AaveAsset), isAave)
		for _, id := range cfg.AaveStrategyIDs {
			routes[id] = yr
		}
		live := cfg.AaveDataProvider != "" && cfg.AaveAsset != ""
		log.Info("perceive: aave supply-APR overlay wired", "strategies", len(cfg.AaveStrategyIDs), "live", live)
	}

	// Morpho: MetaMorpho net APY over the public Blue GraphQL API (no auth).
	// Yield only — Morpho RISK stays with the (auth-gated, currently dormant)
	// Credora adapter. Empty endpoint/vault map => no route, brain sees APY 0.
	if cfg.MorphoAPIEndpoint != "" && len(cfg.MorphoVaults) > 0 {
		yr := morpho.NewYieldReader(nil, cfg.MorphoAPIEndpoint, cfg.ChainID, cfg.MorphoVaults)
		for id := range cfg.MorphoVaults {
			routes[id] = yr
		}
		log.Info("perceive: morpho net-APY overlay wired", "strategies", len(cfg.MorphoVaults))
	}

	if len(routes) == 0 {
		return nil
	}
	return perceive.NewCompositeYieldProvider(routes)
}

// pendleMarketConfigs converts the config's string-addressed Pendle markets into
// the on-chain MarketConfig the pendle reader consumes.
func pendleMarketConfigs(cfg *config.Config) map[types.StrategyID]pendle.MarketConfig {
	out := make(map[types.StrategyID]pendle.MarketConfig, len(cfg.PendleMarkets))
	for id, m := range cfg.PendleMarkets {
		out[id] = pendle.MarketConfig{
			Market:   common.HexToAddress(m.Market),
			PT:       common.HexToAddress(m.PT),
			Duration: m.Duration,
		}
	}
	return out
}

// callerOrNil returns the client as a bind.ContractCaller, or nil when the client
// is nil — so a nil *ethclient.Client does not become a non-nil interface holding
// a nil pointer (which would defeat the adapter's nil-backend no-data guard).
func callerOrNil(client *ethclient.Client) bind.ContractCaller {
	if client == nil {
		return nil
	}
	return client
}

// buildExecutor returns the signing LocalExecutor only when KEEPER_DRY_RUN=false
// (and an RPC + vault + curator key are present); otherwise the dry-run
// LogExecutor. Dry-run is the default and the only path that needs no key.
func buildExecutor(cfg *config.Config, client *ethclient.Client, log *slog.Logger) execute.Executor {
	if cfg.DryRun {
		return execute.NewLogExecutor(log)
	}
	if client == nil || cfg.VaultAddress == "" {
		log.Error("KEEPER_DRY_RUN=false requires KEEPER_RPC_URL and KEEPER_VAULT_ADDRESS")
		os.Exit(1)
	}
	key, err := execute.LoadKey(os.Getenv(cfg.KeyEnv))
	if err != nil {
		log.Error("load curator key", "env", cfg.KeyEnv, "err", err)
		os.Exit(1)
	}
	e, err := execute.NewLocalExecutor(client, common.HexToAddress(cfg.VaultAddress), key, big.NewInt(cfg.ChainID), log)
	if err != nil {
		log.Error("build LocalExecutor", "err", err)
		os.Exit(1)
	}
	log.Warn("execute: LIVE LocalExecutor — transactions will be signed and sent", "vault", cfg.VaultAddress)
	return e
}
