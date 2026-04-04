# Infrastructure

This project has two hostable services:

- `observer/`: the SvelteKit web UI used by observers
- `server/`: the Go WebSocket/API server used by the relay and observer

The `relay/` app is not intended to be centrally hosted. It is run locally by each runner on their own machine.

## Observer

The observer app supports three deployment styles:

- Cloudflare deployment via Wrangler and `@sveltejs/adapter-cloudflare`
- Docker via `observer/Dockerfile` and `observer/docker-compose.yml`
- Manual process hosting using the package scripts

### Environment variables

See `observer/.env.example`:

```env
# The display name of the site
PUBLIC_SITE_NAME="NinjaParty"

# Server host:port (no protocol)
PUBLIC_SERVER_URL=localhost:8080
# Use wss/https instead of ws/http for WebSocket and API connections. this is default true.
# PUBLIC_USE_SECURE_URLS=false
```

### Cloudflare deployment

The Cloudflare deployment is configured by `observer/wrangler.toml`.

1. Install dependencies:

```bash
cd observer
bun install
```

2. Review `wrangler.toml` and update any deployment-specific values:

- `name`: the Worker name
- `[[routes]] pattern`: your custom domain, if you are using one
- `workers_dev`: set to `true` if you want a `*.workers.dev` URL

3. Set the public runtime variables for the Worker.

Recommended: set them in the Cloudflare dashboard for the Worker under Variables and Secrets.

Alternative: define them in `wrangler.toml` under `[vars]`. Cloudflare environment variable docs:

`https://developers.cloudflare.com/workers/configuration/environment-variables/`

The checked-in `wrangler.toml` uses `keep_vars = true`, which allows dashboard-managed variables to survive deploys.

4. Build and deploy:

```bash
cd observer
bun run deploy
```

This runs the Cloudflare build and then `wrangler deploy`.

### Observer on Cloudflare with NinjaParty values

If you are deploying the public NinjaParty setup, use:

```text
PUBLIC_SITE_NAME=NinjaParty
PUBLIC_SERVER_URL=np-server.trifall.com
PUBLIC_USE_SECURE_URLS=true
```

### Docker deployment

The Docker image builds the observer with the Node adapter and runs it on port `3000`.

1. Create `observer/.env` from `observer/.env.example` and adjust values.
2. Start the container:

```bash
cd observer
docker compose up -d --build
```

Notes:

- `observer/docker-compose.yml` exposes the app on `3000`
- `ORIGIN` should match the public URL used to access the observer UI
- `PUBLIC_SERVER_URL` must point at your hosted `server/` instance
- `PUBLIC_USE_SECURE_URLS` should usually be `true` behind HTTPS

### Manual process deployment

You can also run the observer without Docker.

1. Create `observer/.env` from `observer/.env.example`.
2. Install dependencies:

```bash
cd observer
bun install
```

3. Build the Node version:

```bash
bun run build:node
```

4. Start the app:

```bash
bun run start
```

Useful scripts:

- `bun run dev`: local development
- `bun run build:node`: production build for the Node adapter
- `bun run build:cloudflare`: production build for Cloudflare
- `bun run start`: run the built Node server
- `bun run deploy`: build for Cloudflare and deploy with Wrangler

## Server

The server is a Go application that exposes the WebSocket and HTTP API used by the relay and observer.

It supports two deployment styles:

- Docker via `server/Dockerfile` and `server/docker-compose.yml`
- Manual execution as a compiled Go binary

### Environment variables

See `server/.env.example`:

```env
OBSERVER_URL="https://np.trifall.com"
```

Additional supported settings from the current server configuration:

```env
LISTEN_ADDR=":8080"
ROOM_TTL="2h"
DEBUG=0
```

Notes:

- `OBSERVER_URL` should be the base URL of the observer app
- `LISTEN_ADDR` controls the bind address and port
- `ROOM_TTL` controls how long inactive rooms remain in memory
- `DEBUG=1` enables development-style logging

### Docker deployment

Docker is the simplest self-hosting path for the server.

1. Create `server/.env` and set at least `OBSERVER_URL`.
2. Start the container:

```bash
cd server
docker compose up -d --build
```

The default compose setup publishes container port `8080` to host port `8080`.

### Manual binary deployment

1. Create `server/.env` if you want to configure the server without shell exports.
2. Build the binary:

```bash
cd server
go build -o NinjaParty-Server .
```

3. Run it:

```bash
./NinjaParty-Server
```

You can also run it directly without a separate build step:

```bash
go run .
```

## Configuration precedence

In practice, the hostable services are configured in this order:

1. Environment variables from the process or container runtime
2. `.env` files in the service directory, where supported
3. Application defaults

For the Go server, configuration is loaded with [viper](https://github.com/spf13/viper) and `.env` support via [godotenv](https://github.com/joho/godotenv).
