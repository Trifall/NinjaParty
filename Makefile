.PHONY: dev-relay dev-server dev-observer build-relay build-server build-observer test clean

RELAY_DIR := relay
SERVER_DIR := server
OBSERVER_DIR := observer

dev-relay:
	cd $(RELAY_DIR) && DEBUG=1 TLS_SKIP_VERIFY=true go run .

dev-server:
	cd $(SERVER_DIR) && DEBUG=1 go run .

dev-observer:
	cd $(OBSERVER_DIR) && bun run dev

build-relay:
	goreleaser build --snapshot --clean --single-target -f relay/.goreleaser.yaml

build-server:
	goreleaser build --snapshot --clean --single-target -f server/.goreleaser.yaml

build-observer:
	cd $(OBSERVER_DIR) && bun run build

test:
	cd $(RELAY_DIR) && go test -v ./... || true && golangci-lint run ./... || true
	cd $(SERVER_DIR) && go test -v ./... || true && golangci-lint run ./... || true
	cd $(OBSERVER_DIR) && bun run test

clean:
	rm -rf $(RELAY_DIR)/dist
	rm -rf $(SERVER_DIR)/dist
	rm -rf $(OBSERVER_DIR)/.svelte-kit
	rm -rf $(OBSERVER_DIR)/build

install-deps:
	cd $(OBSERVER_DIR) && bun install

lint:
	cd $(OBSERVER_DIR) && bun run lint

format:
	cd $(OBSERVER_DIR) && bun run format:fix
