#!/usr/bin/env bash
# Deletes the local kind cluster: every pod, Traefik, and all database data.
# Docker images and k8s/secrets.env stay, so `pnpm k8s:up` rebuilds it all.
#
# Usage:
#   pnpm k8s:down         # asks first
#   pnpm k8s:down --yes   # no question (scripts, CI)
set -euo pipefail

CLUSTER=fishertimer

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  echo "No kind cluster '$CLUSTER', nothing to delete."
  exit 0
fi

if [ "${1:-}" != "--yes" ]; then
  printf "Delete cluster '%s' and ALL its database data? [y/N] " "$CLUSTER"
  read -r answer
  case "$answer" in
    y|Y|yes|YES) ;;
    *) echo "Cancelled."; exit 0 ;;
  esac
fi

kind delete cluster --name "$CLUSTER"
echo "Deleted. Bring it back with: pnpm k8s:up"
