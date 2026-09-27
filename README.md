# AIR

A minimal self-hosted OTA update server for Expo / React Native apps.
It implements the Expo Updates protocol, so the app uses `expo-updates` as the client.

- `server/` — Go backend (Postgres for metadata, Cloudflare R2 for bundles and assets)
- `server/web/` — dashboard (React + Vite), embedded into the Go binary
- `cli/` — TypeScript CLI for publishing updates
- `example-app/` — Expo app for testing

## Quick start

```bash
cp server/.env.example server/.env
make up           # start Postgres
make migrate-up   # create tables
make create-admin email=you@example.com name="Your Name"   # prints the password
make web-install  # dashboard dependencies (needs bun)
make web-build    # build the dashboard into server/web/dist
make run          # API + dashboard on http://localhost:8080
```

Dashboard development with hot reload: run `make run` and `make web-dev`,
then open http://localhost:5173 (Vite proxies `/api` to :8080).


