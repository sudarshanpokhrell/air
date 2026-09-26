# AIR

A minimal self-hosted OTA update server for Expo / React Native apps.
It implements the Expo Updates protocol, so the app uses `expo-updates` as the client.

- `server/` — Go backend (Postgres for metadata, Cloudflare R2 for bundles and assets)
- `cli/` — TypeScript CLI for publishing updates
- `example-app/` — Expo app for testing

## Quick start

```bash
cp server/.env.example server/.env
make up           # start Postgres
make migrate-up   # create tables
make create-admin email=you@example.com name="Your Name"   # prints the password
make run          # start the API on :8080
curl localhost:8080/api/v1/health
```


