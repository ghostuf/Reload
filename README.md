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
| `FRONTEND_URL` | backend | The frontend's public URL, used to build the `success_url`/`failure_url` eSewa redirects back to. Usually the same value as `FRONTEND_ORIGIN` — kept separate because one is a CORS match and the other is a redirect target |
| `PAYMENT_MODE` | backend | `mock` or `esewa`. Defaults to `esewa` whenever `ESEWA_SECRET_KEY` is set (which it is, by default) — set to `mock` to fall back to the no-payment-processor flow if the sandbox is unreachable during a demo |
| `ESEWA_ENV` | backend | `sandbox` (default) or `production`. Only picks the default form/status URLs below when `ESEWA_FORM_URL`/`ESEWA_STATUS_URL` are left unset |
| `ESEWA_PRODUCT_CODE` | backend | Merchant code eSewa issues. Defaults to eSewa's own fixed UAT code, `EPAYTEST` |
| `ESEWA_SECRET_KEY` | backend | Signs outgoing requests and verifies eSewa's responses. **Never sent to the frontend.** Defaults to eSewa's own published sandbox key — a production deployment must replace it with the real merchant secret eSewa issues |
| `ESEWA_FORM_URL` | backend | Where the frontend's hidden form POSTs to. Defaults based on `ESEWA_ENV` |
| `ESEWA_STATUS_URL` | backend | eSewa's transaction status-check API, called server-to-server to independently confirm a payment. Defaults based on `ESEWA_ENV` |
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
| GET | `/config/payment` | Reports the active `mode` (`mock` or `esewa`) so the frontend knows which checkout flow to use |
| POST | `/checkout` | 🔒 the mock flow: turns the cart into a *completed* order and library entries immediately. Disabled (400) unless `PAYMENT_MODE=mock` |
| POST | `/checkout/esewa/initiate` | 🔒 prices the cart server-side, creates a *pending* order and payment, returns the signed form fields to POST to eSewa. Disabled unless `PAYMENT_MODE=esewa` |
| POST | `/checkout/esewa/verify` | 🔒 body `{data}` — the base64 payload from `success_url`. Verifies the signature, independently re-confirms with eSewa, then completes the order. Idempotent |
| POST | `/checkout/esewa/failure` | 🔒 body `{transaction_uuid}` — marks a pending payment cancelled; the cart is left untouched |
| GET | `/orders` | 🔒 history, newest first — includes `status`, and `payment_status`/`provider_ref_id` for eSewa orders |
| GET | `/orders/:id` | 🔒 full receipt — includes the `payment` object for eSewa orders |
| GET | `/library` | 🔒 owned games |
| PUT/DELETE | `/reviews/:id` | 🔒 author only (admins can additionally delete any review) |
| POST | `/admin/games` | 🛡️ create a game, its genres, and gallery images |
| PUT | `/admin/games/:id` | 🛡️ full replace, including discounts |
| DELETE | `/admin/games/:id` | 🛡️ soft delete — existing orders/library entries keep working |

## Payments (eSewa ePay v2)

Checkout goes through [eSewa](https://developer.esewa.com.np/pages/Epay)'s
ePay v2 integration, in sandbox mode by default:

1. **Initiate** — `/cart`'s Checkout button calls `POST
   /checkout/esewa/initiate`. The backend prices the cart itself (never
   trusting a client-supplied amount), creates a *pending* order and
   payment, and returns a signed set of form fields. The frontend builds a
   hidden form from them and submits it — a real browser POST, leaving the
   app entirely for eSewa's own login page.
2. **Pay** — log in with one of eSewa's own published sandbox wallets:

   | eSewa ID | Password | MPIN | Token |
   |---|---|---|---|
   | `9806800001` (through `...005`) | `Nepal@123` | `1122` | `123456` |

   (The cart page shows this same note automatically whenever `PAYMENT_MODE=esewa`.)
3. **Verify** — eSewa redirects to `/checkout/esewa/success?data=<base64>`.
   That page calls `POST /checkout/esewa/verify`, which:
   - recomputes the response's HMAC-SHA256 signature from its own
     `signed_field_names` and compares it in constant time;
   - **independently** calls eSewa's transaction status-check API rather
     than trusting the callback alone — only `COMPLETE` counts, and a
     completed-but-wrong-amount response is treated as the fraud case
     eSewa's own docs call out, not honoured;
   - on success, marks the order and payment completed, grants the
     library, and clears the cart, all in one transaction;
   - is idempotent — refreshing the success page re-sends the same
     verified data and gets the same order back without re-granting
     anything.
4. **Cancel/failure** — landing on `/checkout/esewa/failure` (eSewa doesn't
   document anything appended to this URL, so the backend embeds
   `transaction_uuid` in its own query string beforehand) marks the pending
   payment cancelled and leaves the cart untouched, so the user can retry.
5. **Abandoned sessions** — a pending order older than about 30 minutes is
   lazily cancelled the next time a user's orders are read; there's no
   background job.

Set `PAYMENT_MODE=mock` (in `.env`) to fall back to the no-payment-processor
`POST /checkout` flow instantly, if the sandbox is unreachable mid-demo.

> **Currency note:** eSewa only moves Nepali Rupees, and every amount this
> integration sends and checks is implicitly NPR. The existing storefront UI
> (`lib/pricing.ts`'s `formatPrice`, plus a few components that format prices
> inline) prints a `$` sign everywhere, inherited from before this feature
> existed. That's a pre-existing display bug this change didn't introduce
> and didn't fix — flagging it rather than changing currency formatting
> across the whole site, which is out of scope here.

## Demo script

A walkthrough to run in class, end to end:

1. **Browse** the storefront at `/` — search, filter by genre, sort.
2. **Register** a new account.
3. Open a game, **add it to the cart**.
4. Go to `/cart` and **check out** — note the sandbox credentials box, the
   "Redirecting to eSewa…" state, and the real redirect to eSewa's login
   page. Log in with one of the sandbox wallets above.
5. Land back on `/checkout/esewa/success`, watch it verify, and land on the
   receipt — note the payment status and eSewa reference code.
6. Visit `/library` — the purchased game is there.
7. Back on the game's page, **leave a review** (this only works because you
   now own it — try it on a game you *haven't* bought to show the 403).
8. **Log in as the admin** account (`make promote-admin EMAIL=...` ahead of
   time), go to `/admin/games`, open a game, and **edit its discount** —
   then reload the storefront and point out the updated price and the
   `on_sale` filter picking it up.
9. Optional: start a second checkout and click away instead of paying, to
   show `/checkout/esewa/failure` and the order landing as "cancelled" in
   `/orders`.

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
   - `FRONTEND_URL` → the same real production frontend URL as
     `FRONTEND_ORIGIN` — eSewa's `success_url`/`failure_url` are built from
     this, so a stale value here sends real payers back to the wrong place
   - `ESEWA_ENV=production`, plus the real `ESEWA_PRODUCT_CODE` and
     `ESEWA_SECRET_KEY` eSewa issues after onboarding — the sandbox values
     this defaults to do not work against eSewa's live system
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
