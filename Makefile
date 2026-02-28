.PHONY: dev-relay dev-server dev-observer build-relay build-server build-observer test clean

RELAY_DIR := relay
SERVER_DIR := server
OBSERVER_DIR := observer

dev-relay:
	cd $(RELAY_DIR) && TLS_SKIP_VERIFY=true go run .

dev-server:
	cd $(SERVER_DIR) && go run .

dev-observer:
	cd $(OBSERVER_DIR) && bun run dev

build-relay:
	cd $(RELAY_DIR) && go build -o bin/NinjaParty-Relay .

build-server:
	cd $(SERVER_DIR) && go build -o bin/NinjaParty-Server .

build-observer:
	cd $(OBSERVER_DIR) && bun run build

test:
	cd $(RELAY_DIR) && go test ./... || true
	cd $(SERVER_DIR) && go test ./... || true
	cd $(OBSERVER_DIR) && bun run test

clean:
	rm -rf $(RELAY_DIR)/bin
	rm -rf $(SERVER_DIR)/bin
	rm -rf $(OBSERVER_DIR)/.svelte-kit
	rm -rf $(OBSERVER_DIR)/build

install-deps:
	cd $(OBSERVER_DIR) && bun install

lint:
	cd $(OBSERVER_DIR) && bun run lint

format:
	cd $(OBSERVER_DIR) && bun run format:fix
