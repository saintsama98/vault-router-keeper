// Package chain wraps the Ethereum RPC connection shared by the perceive and
// execute layers.
package chain

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Dial opens a JSON-RPC connection. The returned *ethclient.Client satisfies
// bind.ContractBackend, so the same handle backs both the perceive ChainReader
// (read-only eth_call) and the execute LocalExecutor (signed transactions).
func Dial(ctx context.Context, rpcURL string) (*ethclient.Client, error) {
	return ethclient.DialContext(ctx, rpcURL)
}
