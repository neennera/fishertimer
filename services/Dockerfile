# One Dockerfile for every Go service. Build from the repo root:
#   docker build -f services/Dockerfile --build-arg SERVICE=account -t fishertimer/account .
# The repo root is the context because each go.mod replaces
# pkg/events and proto with ../../ paths (go.work ties them together).

# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src

# Download modules first so this layer is cached between code changes.
# The cache mounts keep Go's module and compile caches between builds, shared
# by all 7 services, so common libraries are downloaded and compiled once.
COPY go.work go.work.sum ./
COPY pkg/events/go.mod pkg/events/go.sum* pkg/events/
COPY proto/go.mod proto/go.sum* proto/
COPY services/account/go.mod services/account/go.sum* services/account/
COPY services/admin/go.mod services/admin/go.sum* services/admin/
COPY services/api-gateway/go.mod services/api-gateway/go.sum* services/api-gateway/
COPY services/leaderboard/go.mod services/leaderboard/go.sum* services/leaderboard/
COPY services/reward/go.mod services/reward/go.sum* services/reward/
COPY services/study-session/go.mod services/study-session/go.sum* services/study-session/
COPY services/study-timer/go.mod services/study-timer/go.sum* services/study-timer/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Probe for Docker HEALTHCHECKs: the final image has no shell, curl or wget.
# Built before the SERVICE arg so every service shares this cached layer.
COPY tools/healthcheck ./tools/healthcheck
RUN cd tools/healthcheck && GOWORK=off CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/healthcheck .

COPY pkg ./pkg
COPY proto ./proto
COPY services ./services

# Declared here, not at the top: an ARG is part of the cache key of every RUN
# after it, so declaring it earlier re-ran `go mod download` for each service.
ARG SERVICE
# Static binary: no C libraries needed at runtime.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    test -n "$SERVICE" || (echo "pass --build-arg SERVICE=<name>" && exit 1) \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/server ./services/${SERVICE}/cmd

# ---- run stage ----
# distroless/static: just CA certs + tzdata, no shell, runs as non-root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
COPY --from=build /out/healthcheck /healthcheck
ENTRYPOINT ["/server"]
