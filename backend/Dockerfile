# syntax=docker/dockerfile:1

# ---- build stage -----------------------------------------------------------
FROM golang:1.25-alpine AS build

WORKDIR /src

# Dependencies first so they are cached across rebuilds that only touch
# application code.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produces a fully static binary -- gorm's postgres driver is
# pure Go (jackc/pgx), so nothing here needs cgo -- which is what lets the
# final stage be distroless with no libc at all.
ENV CGO_ENABLED=0 GOOS=linux

# All three entrypoints land in the same image: the server, plus the seed and
# promote-admin one-off commands, so `docker compose run --rm backend ./seed`
# works without a second image to build or a Go toolchain on the host.
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed && \
    go build -trimpath -ldflags="-s -w" -o /out/promote-admin ./cmd/promote-admin

# ---- final stage ------------------------------------------------------------
# distroless/static: no shell, no package manager, nothing beyond the CA
# certificates and tzdata a Go binary might need -- about as small an attack
# surface as a container gets. The ":nonroot" tag also already runs as a
# non-root user (uid 65532), rather than requiring a USER directive here.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/api /out/seed /out/promote-admin ./

EXPOSE 8080

# distroless has no shell and no curl, so the healthcheck is the binary
# checking itself -- see selfCheck in cmd/api/main.go.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/api", "-healthcheck"]

# CMD rather than ENTRYPOINT, deliberately: an ENTRYPOINT would be prepended
# to any command `docker compose run` is given instead of replaced, so
# `docker compose run --rm backend ./seed` would actually execute
# `/app/api ./seed` -- starting the server with a stray argument rather than
# running the seed binary at all. Plain CMD lets a one-off command, like
# ./seed or ./promote-admin, cleanly replace it instead.
CMD ["/app/api"]
