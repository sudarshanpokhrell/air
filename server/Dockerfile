# AIR server image: the API with the dashboard embedded, plus the migrate
# and admin binaries. Build from server/:
#
#   docker build -t air-server ./server

# 1. Dashboard → web/dist
FROM oven/bun:1 AS web
WORKDIR /web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# 2. Go binaries. The dashboard is embedded with go:embed, so dist must be
#    in place before `go build`.
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/admin ./cmd/admin

# 3. Runtime: no shell, runs as a non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/ /app/
# cmd/migrate reads file://migrations relative to the working directory.
COPY migrations/ /app/migrations/
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/api"]
