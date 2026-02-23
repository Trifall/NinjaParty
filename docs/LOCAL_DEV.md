# Local Development Guide

## Quick Start

```bash
# Install observer dependencies
make install-deps

# Run the observer (SvelteKit frontend)
make dev-observer

# Run the central server (after Phase 2)
make dev-server

# Run the local relay (after Phase 3)
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

## Windows Defender Warning

### Important: Unsigned Executable

The relay binary (`Ninjabrain-Relay.exe`) is an **unsigned executable**. When you first run it on Windows, you will likely encounter one of these warnings:

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
5. Select the `Ninjabrain-Relay.exe` file

### Why This Happens

- The executable is not code-signed with a Microsoft-trusted certificate
- It opens network ports (WebSocket connections)
- Go binaries sometimes trigger heuristics due to their static linking

## Environment Variables

### Observer (.env)

```bash
PUBLIC_SITE_NAME="Ninjabrain Party Relay"
PUBLIC_API_URL=http://localhost:8080
PUBLIC_WS_URL=ws://localhost:8080
```

### Server

```bash
LISTEN_ADDR=:8080
UI_DOMAIN=http://localhost:5173
ROOM_TTL=2h
```

## Development Ports

| Service | Port |
|---------|------|
| Observer (SvelteKit) | 5173 |
| Central Server | 8080 |
| Local Relay UI | 8080 |
| Ninjabrain Bot API | 52533 |
