# syntax=docker/dockerfile:1

# ---- deps stage -------------------------------------------------------------
# Dependencies in their own stage and keyed only off the lockfile, so editing
# application code never invalidates this layer's cache.
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# ---- build stage -------------------------------------------------------------
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* variables are inlined into the client bundle at build time --
# there is no later point at which they could still be injected -- so this
# one has to arrive as a build arg, not a runtime environment variable.
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

RUN npm run build

# ---- runner stage -------------------------------------------------------------
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
# The standalone server binds to localhost by default; without this it is
# unreachable from any other container, including a reverse proxy.
ENV HOSTNAME=0.0.0.0
ENV PORT=3000

RUN addgroup -S nodejs && adduser -S nextjs -G nodejs

# output: "standalone" (next.config.ts) traces only the files next's own
# server needs and prunes node_modules accordingly; public/ and .next/static
# are deliberately left out of it (meant to be served by a CDN) and are
# copied in by hand here instead.
COPY --from=build --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=build --chown=nextjs:nodejs /app/public ./public

USER nextjs

EXPOSE 3000

# No shell and no curl in this image's minimal server runtime path either, so
# the healthcheck is a one-line Node script hitting the app's own port.
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD ["node", "-e", "fetch('http://127.0.0.1:' + (process.env.PORT || 3000)).then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"]

CMD ["node", "server.js"]
