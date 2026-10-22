#!/usr/bin/env bash
# Brings up the whole app on a local kind cluster, from nothing or from where
# it left off. Safe to run again: every step skips or updates what exists.
#
# Usage (from anywhere in the repo):
#   pnpm k8s:up
#
# Then open http://fishertimer.localhost
#
# Steps, each reusing the repo's own command so versions live in one place:
#   1. check tools and k8s/secrets.env
#   2. create the kind cluster (k8s/kind-config.yaml) if it doesn't exist
#   3. pnpm k8s:traefik  - Gateway API CRDs + Traefik (the front door)
#   4. pnpm images       - build all images and load them into the cluster
#   5. pnpm k8s:apply    - every manifest in k8s/
#   6. wait until every pod is ready
set -euo pipefail

CLUSTER=fishertimer
CONTEXT="kind-$CLUSTER"
NAMESPACE=fishertimer
URL=http://fishertimer.localhost

cd "$(dirname "$0")/.."

step() { printf '\n==> %s\n' "$*"; }
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

# ---- 1. prerequisites -------------------------------------------------------
step "checking tools"
for tool in docker kind kubectl helm pnpm; do
  command -v "$tool" >/dev/null || fail "$tool is not installed (brew install $tool)"
done
docker info >/dev/null 2>&1 || fail "Docker is not running (start OrbStack or Docker Desktop)"

step "checking k8s/secrets.env"
[ -f k8s/secrets.env ] || fail "k8s/secrets.env is missing.
  cp k8s/secrets.env.example k8s/secrets.env
  then replace every change-me / your-... value (see README, Kubernetes section)"
if grep -qE '=(change-me|your-)' k8s/secrets.env; then
  fail "k8s/secrets.env still has placeholder values (change-me / your-...)"
fi

# ---- 2. cluster -------------------------------------------------------------
if kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  step "cluster '$CLUSTER' already exists, reusing it"
  existed=true
else
  step "creating cluster '$CLUSTER'"
  kind create cluster --config k8s/kind-config.yaml
  existed=false
fi

# ---- 3. front door ----------------------------------------------------------
step "installing Gateway API CRDs and Traefik"
pnpm k8s:traefik

# ---- 4. images --------------------------------------------------------------
step "building images and loading them into the cluster"
pnpm images

# ---- 5. manifests -----------------------------------------------------------
step "applying k8s/"
pnpm k8s:apply

# Images are tagged :latest, so a rebuilt image doesn't change the pod spec and
# running pods would keep the old one. Restart them to pick up the new images.
if [ "$existed" = true ]; then
  step "restarting deployments to use the freshly built images"
  kubectl --context "$CONTEXT" -n "$NAMESPACE" rollout restart deployment
fi

# ---- 6. wait ----------------------------------------------------------------
step "waiting for every pod to be ready (first run downloads DB images, a few minutes)"
# Give new pods a moment to be created before waiting on them.
sleep 5
kubectl --context "$CONTEXT" -n "$NAMESPACE" rollout status deployment --timeout=10m
for sts in $(kubectl --context "$CONTEXT" -n "$NAMESPACE" get statefulset -o name); do
  kubectl --context "$CONTEXT" -n "$NAMESPACE" rollout status "$sts" --timeout=10m
done

kubectl --context "$CONTEXT" -n "$NAMESPACE" get pods
printf '\nFisher Timer is up: %s\n' "$URL"
