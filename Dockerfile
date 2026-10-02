

# Dashboard
FROM --platform=$BUILDPLATFORM oven/bun:1 AS web
WORKDIR /src/web
COPY server/web/package.json server/web/bun.lock ./
RUN bun install --frozen-lockfile
COPY server/web/ ./
RUN bun run build

#Go binaries
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .

COPY --from=web /src/web/dist ./web/dist
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && go build -trimpath -ldflags="-s -w" -o /out/admin ./cmd/admin

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 air
WORKDIR /app
COPY --from=go /out/ /usr/local/bin/

COPY server/migrations ./migrations
USER air
ENV PORT=8080 ENV=production
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/health >/dev/null || exit 1
CMD ["sh", "-c", "migrate up && exec api"]
