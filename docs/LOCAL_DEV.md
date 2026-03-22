# Local Development Guide

## Quick Start

```bash
# Install observer dependencies
make install-deps

# Run the observer (SvelteKit frontend)
make dev-observer

# Run the central server
make dev-server

# Run the local relay
make dev-relay
```

## Project Structure

```
/ (root)
├── relay/          # Go module: Local player relay
├── server/         # Go module: Central headless API (REST + WS)
├── observer/       # SvelteKit: Coach web UI
├── shared/         # JSON Schema & example payloads
├── infra/          # Docker configs for deployment
└── Makefile        # Dev commands
```

## Requirements

- **Go 1.21+** for relay and server
- **Bun** for observer frontend
- **Docker** for deployment (optional)

## Prerequisites

### Enable Ninjabrain API Server

1. Open Ninjabrain Bot
2. Go to **Settings** → **Advanced** tab
3. Check **"Enable API (starts HTTP server)"**

This starts the HTTP server that the relay connects to for SSE data.

## Windows Defender Warning

### Important: Unsigned Executable

The relay binary (`NinjaParty-Relay.exe`) is an **unsigned executable**. When you first run it on Windows, you will likely encounter one of these warnings:

1. **Windows SmartScreen** - "Windows protected your PC"
2. **Windows Defender** - Potential false positive detection

### How to Bypass SmartScreen

1. Click **"More info"**
2. Click **"Run anyway"**

### How to Whitelist in Windows Defender

If Windows Defender blocks the executable:

1. Open **Windows Security** → **Virus & threat protection**
2. Click **"Manage settings"** under Virus & threat protection settings
3. Scroll to **Exclusions** and click **"Add or remove exclusions"**
4. Click **"Add an exclusion"** → **"File"**
5. Select the `NinjaParty-Relay.exe` file

### Why This Happens

- The executable is not code-signed with a Microsoft-trusted certificate
- It opens network ports (WebSocket connections)
- Go binaries sometimes trigger heuristics due to their static linking

## Environment Variables

Configuration uses [viper](https://github.com/spf13/viper) with support for `.env` files via [godotenv](https://github.com/joho/godotenv).

### Precedence Order

1. **Environment variables** - Actual env vars set in your shell
2. **`.env` file** - Loaded from the service directory (e.g., `relay/.env`, `server/.env`)
3. **Defaults** - Hardcoded defaults in the code

### Observer (.env)

```bash
PUBLIC_SITE_NAME="NinjaParty"
PUBLIC_SERVER_URL=localhost:8080
# True = wss/https instead of ws/http for WebSocket and API connections. This is default to true in production.
# PUBLIC_USE_SECURE_URLS=false
```

### Server

```bash
LISTEN_ADDR=:8080
OBSERVER_URL=http://localhost:5173
ROOM_TTL=2h
```

### Local Relay

```bash
TLS_SKIP_VERIFY=true  # Skip TLS verification for local development
```

The relay is started via `make dev-relay` which automatically sets `TLS_SKIP_VERIFY=true` to allow local connections and avoid certificate errors.

## Development Ports

| Service | Port |
|---------|------|
| Observer (SvelteKit) | 5173 |
| Central Server | 8080 |
| Local Relay UI | 52532 |
| Ninjabrain Bot API | 52533 |
