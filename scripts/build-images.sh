#!/usr/bin/env bash
# Builds the Docker image for every app (7 Go services + web) and, if a kind
# cluster named "fishertimer" is running, loads them into it.
#
# Usage (from anywhere in the repo):
#   pnpm images                     # build all 8
#   pnpm images reward web          # build only these
#   ./scripts/build-images.sh account
#
# Images are tagged fishertimer/<name>:latest.
set -euo pipefail

ALL_SERVICES=(account admin api-gateway leaderboard reward study-session study-timer)
CLUSTER=fishertimer

cd "$(dirname "$0")/.."

targets=("$@")
if [ ${#targets[@]} -eq 0 ]; then
  targets=("${ALL_SERVICES[@]}" web)
fi

# Fail before building anything if a name is misspelled.
for t in "${targets[@]}"; do
  if [ "$t" != web ] && [[ ! " ${ALL_SERVICES[*]} " =~ " $t " ]]; then
    echo "unknown image '$t'. Choose from: ${ALL_SERVICES[*]} web" >&2
    exit 1
  fi
done

for t in "${targets[@]}"; do
  echo "==> building fishertimer/$t"
  if [ "$t" = web ]; then
    docker build -f apps/web/Dockerfile -t fishertimer/web .
  else
    docker build -f services/Dockerfile --build-arg SERVICE="$t" -t "fishertimer/$t" .
  fi
done

# Images built on your Mac are invisible to kind until loaded into it.
if command -v kind >/dev/null && kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  for t in "${targets[@]}"; do
    echo "==> loading fishertimer/$t into kind cluster '$CLUSTER'"
    kind load docker-image "fishertimer/$t" --name "$CLUSTER"
  done
else
  echo "==> no kind cluster '$CLUSTER' found, skipped loading (images are in local Docker only)"
fi

echo "==> done: ${targets[*]}"
