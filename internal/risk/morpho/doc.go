// Package morpho provides the LIVE yield signal for MetaMorpho (ERC-4626
// curated vault) strategies: the vault's current net APY from Morpho's public
// Blue GraphQL API (https://blue-api.morpho.org/graphql — NO auth required,
// unlike the Credora rating feed in internal/risk/credora, which stays the
// designated RISK source for Morpho strategies and remains dormant until an
// API key is provisioned).
//
// Scope note: this package deliberately carries only a perceive.YieldProvider.
// Morpho risk (EL) keeps flowing through credora; until that feed is live the
// brain prices Morpho at the closed-form floor while this adapter at least
// makes its yield visible — without it the water-fill scores Morpho at APY 0
// and starves it to residual budget regardless of its real rate.
//
// Grounding rule (same as every adapter here): no fabricated numbers. The APY
// returned traces 1:1 to the API's `vaultByAddress.state.netApy` (the
// depositor-realized rate net of vault fees); any transport/shape/zero failure
// degrades to ok=false — the honest no-data path.
package morpho
