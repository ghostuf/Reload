# Reload

Reload is a small digital game store: browse and search a catalog, wishlist
and cart games, check out, build a library of what you own, leave reviews on
games you own, and — as an admin — manage the catalog and its discounts.

## Stack

- **Backend** — Go, [Gin](https://github.com/gin-gonic/gin),
  [GORM](https://gorm.io), Postgres. Cookie-based JWT auth (short-lived
  signed access token + rotating opaque refresh token), full-text search,
  soft-deleted games, role-gated admin routes.
- **Frontend** — Next.js (App Router), React, Tailwind. Server Components for
  the public catalog, client-side data for anything that needs the session
  cookie (cart, orders, library, reviews, admin).
- **Database** — Postgres 17.
- **Everything runs via Docker Compose** — Postgres, the API, and the
  frontend, wired together with healthchecks so each one only starts once
  its dependency is actually ready.

## Running everything

```bash
cp .env.example .env
# open .env and at minimum set JWT_ACCESS_SECRET (openssl rand -base64 48)

docker compose up -d --build
# or: make up
```

That builds and starts Postgres, the backend, and the frontend, in that
order — Compose won't start the backend until Postgres's healthcheck passes,
and won't start the frontend until the backend's does. Once it's up:

- Frontend: http://localhost:3000
- Backend: http://localhost:8080 (`/health` should return `{"status":"ok"}`)

The database starts empty. See below to seed it and create an admin.

```bash
make down    # stop everything, keep the data
make logs    # follow every service's logs
make ps      # see what's running and whether it's healthy
```

## Seeding data

```bash
docker compose run --rm backend ./seed
# or: make seed
```

This truncates and re-inserts ten sample games with genres, gallery images,
and a mix of discounted and full-price titles. Safe to re-run any time you
want a clean catalog again.

## Creating the first admin

There's no sign-up flow for admins — promote an existing account instead:

1. Register a normal account through the frontend (or `POST /auth/register`).
2. Run:
   ```bash
   docker compose run --rm backend ./promote-admin you@example.com
   # or: make promote-admin EMAIL=you@example.com
   ```
3. Log out and back in (or just refresh) so the UI picks up the new role —
   the role itself takes effect immediately server-side; nothing is cached
   in the access token.

## Environment variables

All of these live in one `.env` file at the repo root, which
`docker-compose.yml` reads for every service. `.env.example` has the full
list with placeholder values; `.env` itself is gitignored.

| Variable | Used by | Purpose |
|---|---|---|
| `POSTGRES_USER` | postgres, backend | Database role Postgres creates and the backend connects as |
| `POSTGRES_PASSWORD` | postgres, backend | Password for that role — change this for anything beyond a local demo |
| `POSTGRES_DB` | postgres, backend | Database name |
| `DATABASE_URL` | backend | Full Postgres connection string. Inside Compose the host is the service name `postgres`, not `localhost` |
| `PORT` | backend | Port the API listens on (also what gets published to the host) |
| `JWT_ACCESS_SECRET` | backend | Signs access tokens. **Required** — the server refuses to start without it. Generate with `openssl rand -base64 48` |
| `FRONTEND_ORIGIN` | backend | The exact origin allowed to send credentialed (cookie) requests. No wildcard — cookies forbid one |
| `COOKIE_DOMAIN` | backend | Leave empty to scope cookies to the API's own host. Set to a shared apex domain only if frontend and backend are both under it |
| `COOKIE_SECURE` | backend | Must be `true` for any HTTPS deployment, and whenever `COOKIE_SAMESITE=none` |
| `COOKIE_SAMESITE` | backend | `lax` (default, local HTTP dev), `strict`, or `none` (cross-domain production — see Live deploy notes) |
| `NEXT_PUBLIC_API_URL` | frontend (build-time) | The API's address **from the browser's point of view**. Baked into the client bundle at build time |
| `INTERNAL_API_URL` | frontend (runtime) | The API's address **from inside the frontend container**, used only by Server Components doing their own fetch. Inside Compose this is `http://backend:8080` — `localhost` there would mean the frontend container itself |

## Main API routes

All under `http://localhost:8080`. Routes marked 🔒 require the session
cookie (`RequireAuth`); 🛡️ additionally require the `admin` role.

| Method | Path | Notes |
|---|---|---|
| GET | `/health` | Checks DB connectivity too, not just that the process is up |
| GET | `/games` | `search`, `genre`, `on_sale=true`, `sort`, `page`, `page_size` |
| GET | `/games/:id` | Includes `effective_price`, `average_rating`, `review_count` |
| GET | `/games/:id/reviews` | Paginated, newest first |
| POST | `/games/:id/reviews` | 🔒 only if the game is in your library |
| POST | `/auth/register` | |
| POST | `/auth/login` | |
| POST | `/auth/refresh` | Rotates the refresh token |
| POST | `/auth/logout` | |
| GET | `/auth/me` | 🔒 |
| GET/POST | `/cart` | 🔒 |
| DELETE | `/cart/:game_id` | 🔒 |
| GET/POST | `/wishlist` | 🔒 |
| DELETE | `/wishlist/:game_id` | 🔒 |
| POST | `/checkout` | 🔒 turns the cart into an order and library entries |
| GET | `/orders` | 🔒 history, newest first |
| GET | `/orders/:id` | 🔒 full receipt |
| GET | `/library` | 🔒 owned games |
| PUT/DELETE | `/reviews/:id` | 🔒 author only (admins can additionally delete any review) |
| POST | `/admin/games` | 🛡️ create a game, its genres, and gallery images |
| PUT | `/admin/games/:id` | 🛡️ full replace, including discounts |
| DELETE | `/admin/games/:id` | 🛡️ soft delete — existing orders/library entries keep working |

## Demo script

A walkthrough to run in class, end to end:

1. **Browse** the storefront at `/` — search, filter by genre, sort.
2. **Register** a new account.
3. Open a game, **add it to the cart**.
4. Go to `/cart` and **check out** — note the cart badge clear and the
   receipt that follows.
5. Visit `/library` — the purchased game is there.
6. Back on the game's page, **leave a review** (this only works because you
   now own it — try it on a game you *haven't* bought to show the 403).
7. **Log in as the admin** account (`make promote-admin EMAIL=...` ahead of
   time), go to `/admin/games`, open a game, and **edit its discount** —
   then reload the storefront and point out the updated price and the
   `on_sale` filter picking it up.

## Live deploy notes (optional)

The common shape: **frontend on Vercel, backend + Postgres on Fly.io or
Render**, on two different domains. That changes two things Compose doesn't
have to deal with: cookies crossing a real cross-site boundary, and the
frontend's two API URLs pointing at two different real hosts instead of a
container alias.

### Backend + Postgres — Fly.io or Render

1. Provision a managed Postgres instance; note its connection string.
2. Deploy the backend as a Docker image (the existing `backend/Dockerfile`
   works as-is — both platforms build straight from it).
3. Set the same backend env vars as the table above, with production values:
   - `DATABASE_URL` → the managed Postgres connection string (keep
     `sslmode=require` if the provider expects TLS, which both of these do)
   - `JWT_ACCESS_SECRET` → a freshly generated secret, not the dev one
   - `FRONTEND_ORIGIN` → the frontend's real production URL, e.g.
     `https://reload.vercel.app`
   - `COOKIE_SECURE=true`
   - `COOKIE_SAMESITE=none` — **required** once frontend and backend are on
     different domains, or the browser drops the auth cookie on every
     cross-site request the frontend makes
   - `COOKIE_DOMAIN` → leave empty unless you've put both services under a
     shared parent domain
4. Point the platform's healthcheck at `GET /health`.
5. Run the one-off commands against the deployed container once, the same
   way Compose does locally:
   `./seed` and `./promote-admin you@example.com` (`fly ssh console -C` or
   the Render shell, whichever platform you picked).

### Frontend — Vercel

1. Import the repo, set the root directory to `frontend/`.
2. Environment variables:
   - `NEXT_PUBLIC_API_URL` → the backend's real public URL, e.g.
     `https://reload-api.fly.dev`
   - `INTERNAL_API_URL` → the same backend URL. There's no compose network
     once these are two separate hosted services, so the browser-facing and
     server-side addresses collapse to the same value.
3. Deploy. Vercel builds straight from `next.config.ts`'s `output:
   "standalone"` setting without any extra configuration.

### Why SameSite=None specifically

Locally, the frontend and backend are both `localhost` (different ports,
same site as far as `SameSite=Lax` cares), so `lax` works. Once they're on
different real domains, every request the frontend makes to the backend is
cross-site, and `Lax` cookies are not attached to cross-site `fetch()` calls
at all — only `None` is. Browsers additionally refuse `SameSite=None`
without `Secure`, which is why both have to change together.
