# Ninjabrain Party

Ninjabrain Party is an observation tool for [Ninjabrain Bot](https://github.com/Ninjabrain1/NinjabrainBot). It allows Minecraft speedrunners to broadcast their stronghold triangulation data to observers in real-time (ish) through a web dashboard.

![System Architecture](observer/static/images/system_architecture.png)

## Architecture

The system is split into four main components to ensure make it as simple as possible to setup:

1.  **Ninjabrain Bot (Existing)**: The source of truth, running on the player's local machine with its HTTP API enabled (in Settings -> Advanced -> Enable API).
2.  **Relay (`/relay`)**: A simple Go executable that runs on the player's PC. It reads the local SSE stream from Ninjabrain Bot and forwards the data to the central server via WebSockets.
3.  **Central Server (`/server`)**: A Go application acting as a headless message broker. It manages "Rooms" in-memory and routes data from runners to their observers who have joined the room.
4.  **Observer (`/observer`)**: A modern SvelteKit web dashboard where observers can view live data (eye throws, triangulation results, and player position) by just entering a room code.

## Privacy

-   **No Personal Data**: The central server does not collect or store any personal data from observers nor runners.
-   **No Data Retention**: The central server does not store any data after it has been relayed to observers. The Ninjabrain data is stored in memory, once the room closes, it is gone.

## Why a Relay?

A key choice was using a central relay server instead of direct peer-to-peer (P2P) connections between the player and the observer.

-   **Privacy & Security**: By routing data through a central server, the player's IP address is never exposed to observers. Random people joining a room (code leak, or somehow guessed) only see the data, not the player's network location.
-   **Reliability**: Direct connections can fail due to ISP restrictions, NAT traversal issues, or complex router configurations. Using a WebSocket-based relay bypasses these hurdles entirely, requiring zero port-forwarding from the player.
-   **Performance**: The Go-based relay and server provide pretty low latency due to the websocket setup.

## Tech Stack

-   **Backend**: Go (standard library + `chi` for routing, `coder/websocket` for fast communication).
-   **Frontend**: Svelte 5 & SvelteKit. Standard websockets for communication with the server.
-   **Infrastructure**: Dockerized server deployment. Observer (Web UI) is hosted on Cloudflare.

## Repository Structure

```text
├─ relay/      # Player-side local Go relay
├─ server/     # Central Go WebSocket broker
├─ observer/   # SvelteKit web dashboard for observers
├─ shared/     # Shared JSON schemas and types
├─ infra/      # Docker & deployment configurations
└─ Makefile    # Development and build commands
```

## Getting Started / Development

Check out the [LOCAL_DEV.md](docs/LOCAL_DEV.md) for instructions on setting up the project locally.