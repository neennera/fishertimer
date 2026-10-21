# Next.js web app. Build from the repo root:
#   docker build -f apps/web/Dockerfile -t fishertimer/web .
#
# Next.js bakes rewrites() and NEXT_PUBLIC_* in at BUILD time, so the
# in-cluster addresses are passed as build args, not runtime env.

# ---- deps + build stage ----
FROM node:24-alpine AS build
WORKDIR /repo
RUN corepack enable

ARG API_GATEWAY_URL=http://api-gateway:8000
ARG ADMIN_SERVICE_URL=http://admin:8087
ARG NEXT_PUBLIC_APP_URL=http://fishertimer.localhost
ARG NEXT_PUBLIC_API_GATEWAY_URL=http://fishertimer.localhost
ARG NEXT_PUBLIC_ADMIN_SERVICE_URL=http://fishertimer.localhost
ARG NEXT_PUBLIC_APP_NAME="Fisher Timer"
ARG NEXT_PUBLIC_USE_MOCKS=false
ENV API_GATEWAY_URL=$API_GATEWAY_URL \
    ADMIN_SERVICE_URL=$ADMIN_SERVICE_URL \
    NEXT_PUBLIC_APP_URL=$NEXT_PUBLIC_APP_URL \
    NEXT_PUBLIC_API_GATEWAY_URL=$NEXT_PUBLIC_API_GATEWAY_URL \
    NEXT_PUBLIC_ADMIN_SERVICE_URL=$NEXT_PUBLIC_ADMIN_SERVICE_URL \
    NEXT_PUBLIC_APP_NAME=$NEXT_PUBLIC_APP_NAME \
    NEXT_PUBLIC_USE_MOCKS=$NEXT_PUBLIC_USE_MOCKS \
    NEXT_TELEMETRY_DISABLED=1

# Install only what the web app needs (plus its workspace packages).
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY packages ./packages
COPY apps/web/package.json apps/web/
RUN pnpm install --frozen-lockfile --filter web...

COPY apps/web ./apps/web
RUN pnpm --filter web build

# ---- run stage ----
FROM node:24-alpine
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0
# standalone = server.js + only the node_modules it actually imports.
COPY --from=build --chown=node:node /repo/apps/web/.next/standalone ./
COPY --from=build --chown=node:node /repo/apps/web/.next/static ./apps/web/.next/static
COPY --from=build --chown=node:node /repo/apps/web/public ./apps/web/public
USER node
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
