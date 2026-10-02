# AIR

A minimal self-hosted OTA update server for Expo / React Native apps.
It implements the Expo Updates protocol, so the app uses `expo-updates` as the client.

- `server/` — Go backend (Postgres for metadata, Cloudflare R2 for bundles and assets)
- `server/web/` — dashboard (React + Vite), embedded into the Go binary
- `cli/` — TypeScript CLI for publishing updates
- `example-app/` — Expo app for testing

## Quick start

```bash
make dev         
```

That's all the backend needs (Go and Docker). No `.env` file is required: the
server defaults to `PORT=8080`, `ENV=development` and the Postgres from
`make up`. To override any of them, `cp server/.env.example server/.env` and edit it.

First-time extras:

```bash
make create-admin email=you@example.com name="Your Name"   # prints the password
make web-install  # dashboard dependencies (needs bun)
make web-build    # build the dashboard into server/web/dist, then restart `make dev`
```

Dashboard development with hot reload: run `make run` and `make web-dev`,
then open http://localhost:5173 (Vite proxies `/api` to :8080).

## Docker

Run everything in containers — Postgres, migrations, and the API with the
dashboard embedded. Only Docker is required (no Go or bun on the host).

```bash
make docker-up      # build the image, start Postgres, migrate, start the API
make docker-admin email=you@example.com name="Your Name"   # prints the password
```

Then open http://localhost:8080.

| Command             | What it does                                       |
| ------------------- | -------------------------------------------------- |
| `make docker-up`    | Build and start Postgres, migrations, and the API  |
| `make docker-down`  | Stop the containers (the `pgdata` volume is kept)  |
| `make docker-logs`  | Follow the API logs                                |
| `make docker-admin` | Create a global admin and print its password       |

The container applies pending migrations every time it starts, then starts
the API. Run `make docker-up` again after pulling changes to rebuild the image.

To reset a user's password:

```bash
docker compose --profile app exec api admin reset-password --email you@example.com
```

### Configuration

The containers don't read `server/.env`. They are configured through these
variables, set in your shell or in a `.env` file next to `docker-compose.yml`:

| Variable            | Default                 | Description                                  |
| ------------------- | ----------------------- | -------------------------------------------- |
| `POSTGRES_PASSWORD` | `ota`                   | Postgres password (change it in production)  |
| `PUBLIC_URL`        | `http://localhost:8080` | URL the server is publicly reachable at      |
| `AIR_PORT`          | `8080`                  | Host port the API and dashboard are bound to |
| `POSTGRES_PORT`     | `5433`                  | Host port Postgres is published on           |

`POSTGRES_PASSWORD` only takes effect when the database volume is first
created. If you change `POSTGRES_PORT`, set a matching `DATABASE_URL` in `server/.env`
for `make dev`.

### Standalone container

The image is self-contained (API, dashboard, `migrate` and `admin`), so it can
run without Compose against any Postgres. It migrates on startup:

```bash
docker build -t air-server .

docker run -d --name air -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/ota \
  -e PUBLIC_URL=https://ota.example.com air-server

docker exec air admin create-admin --email you@example.com --name "Your Name"
```


