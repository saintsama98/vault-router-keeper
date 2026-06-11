.PHONY: build run tidy fmt vet test test-integration clean bindings

build:
	go build -o bin/keeper ./cmd/keeper

run:
	go run ./cmd/keeper

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

# Self-contained fresh-anvil end-to-end test (build-tag gated so the normal gate
# above never compiles/runs it). Stands up anvil, deploys the diamond via the
# diamond repo's forge script, runs the keeper one tick with a live signing
# executor, and asserts the on-chain targetAllocation changed. Skips cleanly when
# anvil/forge are unavailable. See internal/integration/e2e_test.go RUNBOOK for
# the fresh-anvil vs --fork-url variants.
test-integration:
	GOWORK=off go test -tags integration ./internal/integration/...

clean:
	rm -rf bin

# Regenerate Go contract bindings from raw ABIs in ./abi (requires abigen on PATH).
# ABIs are extracted from the vault-router-diamond Foundry artifacts; see
# abi/README.md for the source mapping.
bindings:
	abigen --abi abi/vault.abi.json         --pkg vault        --type Vault        --out internal/bindings/vault/vault.gen.go
	abigen --abi abi/pendle-oracle.abi.json --pkg pendleoracle --type PendleOracle --out internal/bindings/pendleoracle/pendleoracle.gen.go
	abigen --abi abi/pendle-pt.abi.json     --pkg pendlept     --type PendlePT     --out internal/bindings/pendlept/pendlept.gen.go
	abigen --abi abi/aave-pool.abi.json     --pkg aavepool     --type AavePool     --out internal/bindings/aavepool/aavepool.gen.go
	abigen --abi abi/aave-data.abi.json     --pkg aavedata     --type AaveData     --out internal/bindings/aavedata/aavedata.gen.go
