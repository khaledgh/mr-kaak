# Build Plan — Arabic Sweets & Lebanese Kaak Ordering Platform (Canada)

**Stack:** React + Vite + TypeScript + TailwindCSS v3 + TanStack Query (frontend) · Go + Echo + GORM + MySQL (backend) · Redis (cache + jobs) · Meilisearch Community Edition, self-hosted (search) · Square (payments)

> **Cost note:** every infrastructure component here is free/open-source and self-hosted — MySQL, Redis, and **Meilisearch Community Edition (MIT license, $0)**. The only usage cost is Square's per-transaction processing fee. No paid SaaS, no per-search billing, no vendor lock-in.

This document is written to be executed step-by-step by an AI coding assistant. Each phase is self-contained with concrete deliverables, schema, and acceptance criteria.

---

## 1. Product Scope (from the menu images)

The catalog has **two fundamentally different pricing modes**, which drives the whole data model:

| Group | Example items | Pricing mode | Notes |
|---|---|---|---|
| **Kaak / Manakish** | Kaak plain, Cheese, Cheese Basterma, Knefi in Kaak | **per-unit** | Has size variants (L/S), add-ons, ingredient swaps |
| **Sweets / Cakes** | Baklawa, Knefe, Halawet El Jeben, Maamoul, Sfouf | **per-weight (per kg)** | Ordered by weight; some are **pre-order only** |

Concrete cases the schema must handle (all visible in the menu):

1. **Size variants** — `Kaak plain: L 1.50$ / S 1.00$`.
2. **Add-ons / extras** — `Add Veggies +1.5$`, `Extra Cheese +2$`.
3. **Ingredient swap** — `Replace Akkawi with Halloumi +1$`.
4. **Per-kg items** — every sweet (`35.00$ /kg`), with custom weight input at checkout (e.g. 0.5 kg, 1.25 kg).
5. **Pre-order flag** — `Katayef - Pre-order`, `Knefi plate (By Weight)` need lead time / not instantly available.
6. **Option groups** — Katayef: *Kashta Cream / Kashta Cream (Fried) / Cheese / Walnuts (Fried)*; Knefe: *Cheese or Kashta cream*.
7. **Per-item allergens** — "Contains dairy, nuts, gluten."

> All product names, descriptions, and category names must be **translatable** (see §7).

---

## 2. High-Level Architecture

```
                          ┌─────────────────────────────┐
        Customer PWA ─────►            Nginx              ◄──── Admin SPA
   (React+Vite+TS+Query)  │  (TLS, gzip, static, proxy)  │  (React+Vite+TS+Query)
                          └──────────────┬──────────────┘
                                         │
                              ┌──────────▼──────────┐
                              │   Go API (Echo)     │  stateless, N replicas
                              │  models/handlers/   │
                              │  services/repo      │
                              └───┬───────┬─────┬───┘
                                  │       │     │
                ┌─────────────────┘       │     └──────────────────┐
                │                         │                        │
         ┌──────▼──────┐         ┌────────▼────────┐       ┌────────▼────────┐
         │   MySQL 8    │         │     Redis       │       │   Meilisearch   │
         │ (source of   │         │ cache + Asynq   │       │  (fast search)  │
         │  truth, GORM)│         │ queue + pubsub  │       │  synced from DB │
         └──────────────┘         └─────────────────┘       └─────────────────┘
                                          │
                                  ┌───────▼────────┐
                                  │  Asynq workers │  orders, notifications,
                                  │ (goroutine pool│  payment webhooks,
                                  │   processes)   │  search reindex, emails
                                  └────────────────┘
```

**Key decisions**

- **API is fully stateless** → horizontal scaling is trivial (just add replicas behind Nginx). No in-process session state; auth via JWT.
- **Redis** does triple duty: response/query cache, Asynq job queue, and Pub/Sub for live order-status pushes.
- **Meilisearch Community Edition** is the search engine (sub-50 ms typo-tolerant search) — **free, MIT-licensed, self-hosted** as a single Docker container. MySQL stays the source of truth; Meilisearch is a derived index kept in sync via Asynq jobs. (Typesense is an equally valid free open-source swap-in.)
- **Asynq** (Redis-backed Go job queue) handles everything heavy/async: sending push notifications, payment webhook processing, search reindexing, email receipts — keeping HTTP request latency low.

---

## 3. Backend — Go / Echo / GORM / MySQL

### 3.1 Folder structure

You asked for `models / handlers / services`. Here is the full layout that keeps those three but adds the layers a production app needs:

```
backend/
├── cmd/
│   ├── api/main.go            # HTTP server entrypoint
│   └── worker/main.go         # Asynq worker entrypoint (separate process)
├── internal/
│   ├── config/                # env loading + validation (struct-based)
│   ├── models/                # GORM models (DB schema)
│   ├── repository/            # data access; all DB queries live here
│   ├── services/              # business logic; orchestrates repo + cache + jobs
│   ├── handlers/              # Echo HTTP handlers (thin: parse → call service → respond)
│   ├── middleware/            # auth (JWT), RBAC, rate-limit, request-id, recover, CORS
│   ├── cache/                 # Redis client + cache-aside helpers + key registry
│   ├── search/                # Meilisearch client + index sync
│   ├── jobs/                  # Asynq task types + handlers (worker logic)
│   ├── payment/               # payment provider abstraction + implementations
│   ├── i18n/                  # translation resolution helpers
│   ├── geo/                   # distance / point-in-polygon / zone matching
│   ├── realtime/              # SSE/WS hub for order status
│   └── validator/             # request DTO validation
├── migrations/                # SQL migrations (golang-migrate)
├── pkg/
│   ├── logger/                # structured logging (zerolog/slog)
│   ├── response/              # standard JSON envelope + error codes
│   └── pagination/
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

**Layering rule (enforce strictly):**
`handlers → services → repository → DB`. Handlers never touch GORM directly. Services never write SQL. This keeps caching, search sync, and job dispatch all centralized in `services`.

### 3.2 Request lifecycle (example: get menu)

```
GET /api/v1/menu?lang=ar
  handler.GetMenu
    → service.Menu.GetActiveMenu(ctx, lang)
        → cache.GetOrSet("menu:active:ar", ttl, func() { repo.Menu.LoadActive(ctx) })
        → resolve translations for lang, fall back to default lang
    → response.JSON(200, envelope)
```

### 3.3 Concurrency & load management (your "manage pressure" requirement)

This is the part that lets the app survive a busy Friday-night rush.

**a) Server-level**
- Run the API stateless with **N replicas**; Nginx load-balances. Scale by adding containers, not by tuning one process.
- `http.Server` with `ReadTimeout`, `WriteTimeout`, `IdleTimeout` set. Echo `Recover()` middleware so one panic never takes down a worker.
- **Graceful shutdown**: catch SIGTERM, stop accepting new requests, drain in-flight (with `context` deadline), then close DB/Redis. Critical for zero-downtime deploys.

**b) Context propagation**
- Every handler creates a request `context.Context` with a timeout (e.g. 5 s). Pass it all the way to GORM (`db.WithContext(ctx)`) and Redis. A slow query gets cancelled instead of piling up.

**c) Database connection pool (GORM/`database/sql`)**

```go
sqlDB.SetMaxOpenConns(50)        // tune to MySQL max_connections / replica count
sqlDB.SetMaxIdleConns(25)
sqlDB.SetConnMaxLifetime(30 * time.Minute)
sqlDB.SetConnMaxIdleTime(5 * time.Minute)
```
The pool is your primary backpressure valve: when all connections are busy, requests queue instead of overwhelming MySQL.

**d) Offload everything non-critical to Asynq workers (the "threads to manage big numbers of orders")**

The HTTP request for placing an order should do the *minimum*: validate, reserve, write the order in a transaction, return. Everything else is enqueued:

```
POST /orders  →  (in TX) create order + items, decrement/lock stock if tracked
              →  enqueue: payment.capture, notify.customer, notify.kitchen,
                          email.receipt, search.noop
              →  return 201 with order id immediately
```

Asynq gives you:
- **Worker concurrency** (`asynq.Config{Concurrency: 30}`) = a managed goroutine pool. Tune the number; don't spawn unbounded goroutines.
- **Queues with priorities**: `critical` (payments) > `default` (notifications) > `low` (reindex, analytics).
- **Automatic retries with backoff**, dead-letter handling, and a dashboard (asynqmon).

**e) In-process worker pool pattern (when you need it without Redis)**
For things like fan-out push notifications, use a bounded worker pool so you never launch 10k goroutines at once:

```go
sem := make(chan struct{}, 20) // max 20 concurrent
for _, sub := range subscribers {
    sem <- struct{}{}
    go func(s Sub) {
        defer func() { <-sem }()
        sendPush(s)
    }(sub)
}
```

**f) Rate limiting & abuse protection**
- Echo middleware rate limiter, keyed by IP + user, e.g. 100 req/min general, stricter on `/auth/*` and `/orders` (prevent order spam / coupon brute-forcing).
- Redis-backed sliding window so limits are shared across replicas.

**g) Idempotency for orders & payments**
- Accept an `Idempotency-Key` header on `POST /orders` and payment webhooks. Store processed keys in Redis (24 h TTL) → a retried request never double-charges or double-creates. (This matches your past order-pipeline hardening work — keep zero silent failures, structured logs on every state transition.)

### 3.4 Caching strategy (fast + invalidatable — your explicit requirement)

**Pattern: cache-aside with explicit, surgical invalidation.**

Centralize keys in `cache/keys.go` so invalidation is never guesswork:

```go
MenuActive(lang)      → "menu:active:" + lang          // TTL 10m
Product(id, lang)     → "product:" + id + ":" + lang   // TTL 10m
CategoryList(lang)    → "cat:list:" + lang             // TTL 30m
Banners()             → "banners:active"               // TTL 5m
Settings()            → "settings"                      // TTL 30m
DeliveryZones()       → "geo:zones"                     // TTL 30m
```

**Read path:** `GetOrSet(key, ttl, loaderFn)` — try Redis, on miss load from DB, set, return.

**Write path — "remove caching when I want":** every mutating service method calls a matching invalidator. Two complementary mechanisms:

1. **Targeted delete** — admin edits product 42 → `cache.Del(Product(42, *), MenuActive(*))`. Use Redis key tagging or a small set per entity so you can wipe all language variants at once.
2. **Versioned namespaces (recommended for the menu)** — keep a counter `menu:version`. Cache keys embed it: `menu:v{N}:active:ar`. To invalidate the *entire* menu instantly, just `INCR menu:version` — old keys are orphaned and expire naturally. This makes "flush the whole menu cache" an O(1) operation and avoids race conditions.
3. **Admin "Clear cache" button** → endpoint `POST /admin/cache/flush` (RBAC-protected) that bumps versions / `FLUSHDB` on a dedicated cache DB index. Gives you the manual control you asked for.

**Do NOT cache:** anything user-specific and live (cart, order status, auth). Those read from DB / Redis live keys.

### 3.5 Search (best speed, $0 license)

**Meilisearch Community Edition** as a dedicated index for products + categories — **open source (MIT), self-hosted, free**. Runs as one Docker container (`getmeili/meilisearch`) with a master key and a persistent volume; no per-search fees, no cloud subscription. Deliberately kept as a separate service so it can be restarted/reindexed independently without touching MySQL.

- Index documents: `{id, name_en, name_ar, name_fr, description_*, category, tags, price_from, is_available}`.
- Configure `searchableAttributes`, `filterableAttributes` (category, availability, price), typo tolerance, and synonyms (e.g. "kaak"/"kaake", "knefe"/"knefi/kunafa").
- **Sync via Asynq**: on product create/update/delete, services enqueue a `search.index` / `search.delete` task → worker upserts the doc. MySQL stays authoritative; search is eventually consistent within ~100 ms.
- Frontend calls your Go API (`GET /search?q=...&lang=...`), Go proxies to Meilisearch (never expose the search key to the browser).
- Multi-language: store all language fields in one doc; filter/boost by requested locale.

*Zero-extra-service alternative (also free):* MySQL `FULLTEXT` index + Redis result cache — no second container to run, but no typo tolerance and slower fuzzy matching. Both options cost $0; Meilisearch wins on speed and typo tolerance for a small menu like this, so keep it unless you want to minimize moving parts.

---

## 4. Data Model (MySQL via GORM)

Core tables (abbreviated; add `created_at/updated_at/deleted_at` everywhere for soft deletes):

```
users(id, name, email, phone_e164, password_hash, role, default_address_id, status)
roles/permissions          -- or a role enum: customer|staff|kitchen|admin|super_admin

categories(id, slug, sort_order, image_url, is_active)
products(id, category_id, slug, pricing_mode ENUM('unit','weight'),
         base_price_cents, is_preorder, preorder_lead_hours,
         is_available, image_url, allergens_json, sort_order)

product_variants(id, product_id, sku, label_key, price_cents, is_default)   -- L / S
modifier_groups(id, product_id, label_key, min_select, max_select, is_required)
modifiers(id, group_id, label_key, price_delta_cents, is_default)
                                       -- Extra Cheese +200, Replace Akkawi +100, etc.

translations(id, entity_type, entity_id, locale, field, value)  -- generic i18n store
languages(id, code, name, native_name, is_default, is_rtl, is_active, sort_order)

addresses(id, user_id, label, line1, line2, city, province_code, postal_code,
          country_code 'CA', lat, lng, phone_e164, notes)

delivery_zones(id, name, scope ENUM('global','product'), product_id NULL,
               shape ENUM('radius','polygon'),
               center_lat, center_lng, radius_km,
               polygon_geojson, fee_cents, min_order_cents, is_active)

orders(id, user_id, code, status, subtotal_cents, discount_cents, delivery_fee_cents,
       tax_cents, total_cents, currency 'CAD', payment_method, payment_status,
       coupon_id NULL, address_snapshot_json, scheduled_for NULL, idempotency_key)
order_items(id, order_id, product_id, variant_id, name_snapshot,
            qty, weight_grams NULL, unit_price_cents, modifiers_json, line_total_cents)
order_status_history(id, order_id, from_status, to_status, actor_id, note, created_at)

coupons(id, code, type ENUM('percent','fixed','free_delivery'), value,
        min_order_cents, max_discount_cents, usage_limit, used_count,
        per_user_limit, starts_at, ends_at, is_active)

banners(id, title_key, image_url, link_url, sort_order, starts_at, ends_at, is_active)

payment_transactions(id, order_id, provider, provider_ref, amount_cents,
                     status, raw_payload_json)

settings(key, value_json)   -- languages config, Meta keys, tax %, store hours, etc.
push_subscriptions(id, user_id, endpoint, p256dh, auth, user_agent)
```

**Order status flow** (with `order_status_history` logging every transition):
```
pending_payment → paid → confirmed → preparing → ready → out_for_delivery → delivered
        └──────────────► cancelled / refunded ◄─────────────────────────────┘
(COD orders skip pending_payment: confirmed → preparing → ...)
```

**Money:** store everything as integer **cents** (`*_cents`), never floats. Weight items: store `weight_grams` and compute `line_total = base_price_per_kg * grams / 1000`.

---

## 5. Location Detection & Delivery Zones (by km²)

### 5.1 User location detection (frontend)
- Use the browser **Geolocation API** (`navigator.geolocation.getCurrentPosition`) with a clear permission prompt.
- Reverse-geocode lat/lng → human address using a maps provider (Google Maps / Mapbox / Nominatim) to prefill the address form.
- Fallback: manual address entry with autocomplete; always keep the final `lat/lng` for zone matching.

### 5.2 Zone matching (your "order by km² per item or for all")
This is the geo logic, lives in `internal/geo`:

- **Global zones**: store accepts delivery within radius/polygon X from the shop. `scope='global'`.
- **Per-item zones**: a specific product (e.g. fresh Knefe, or a heavy per-kg cake) may have a *tighter* delivery radius. `scope='product', product_id=...`.
- **Algorithm at checkout:**
  1. Get customer `lat/lng`.
  2. Check the **global** zone first → if outside, reject ("we don't deliver to your area yet").
  3. For each cart item, if it has a product-specific zone, verify the point is inside it → otherwise flag that item as not deliverable to this address.
  4. Compute delivery fee from the matched zone (`fee_cents`, `min_order_cents`).
- **Math:**
  - `radius` shape → **Haversine distance** (great-circle km) between shop and customer; deliverable if `distance ≤ radius_km`.
  - `polygon` shape → **point-in-polygon** (ray casting) against `polygon_geojson` for irregular real-world boundaries.
- Admin draws/edits zones on a map in the dashboard; store as GeoJSON. (Optional later: MySQL spatial types + `ST_Contains` / `ST_Distance_Sphere` if you want the DB to do it — but in-app Go matching is simpler and cache-friendly since zones change rarely and are cached.)

---

## 6. Canada Addressing & Phone

- **Provinces/territories**: enum of 13 codes (AB, BC, MB, NB, NL, NS, NT, NU, ON, PE, QC, SK, YT) with bilingual labels. Seed table.
- **Postal code**: validate `A1A 1A1` format (regex `^[ABCEGHJ-NPRSTVXY]\d[ABCEGHJ-NPRSTV-Z] ?\d[ABCEGHJ-NPRSTV-Z]\d$`), normalize to uppercase with a space.
- **City**: free text with autocomplete (and optional validation against the selected province).
- **Phone**: store **E.164** (`+1XXXXXXXXXX`). Validate Canadian numbers (10 digits, area code). Use `libphonenumber` (Go: `nyaruka/phonenumbers`) for parsing/validation/formatting. (You've done Lebanese normalization before — same approach, locale `CA`.)
- **Country** fixed to `CA` for now; schema leaves room to expand.

---

## 7. Multi-language (admin-controlled, dynamically addable)

**Backend**
- `languages` table is the source of truth — admin adds a language (code, native name, RTL flag, active) from the dashboard; no redeploy needed.
- `translations` table stores per-entity field translations (product names/descriptions, categories, banners, modifier labels).
- Static UI strings: serve a JSON bundle per locale from `GET /i18n/{locale}.json`, also editable in admin (stored in `settings`/`translations`). New language → admin creates the locale, fills strings, toggles active.
- Resolution: requested locale → fall back to `is_default` language for any missing field. Never return an empty label.

**Frontend**
- **i18next + react-i18next**, loading bundles from the API (so adding a language server-side instantly works client-side).
- **RTL**: when `is_rtl` (Arabic), set `dir="rtl"` on `<html>` and use Tailwind logical utilities (`ps-`, `pe-`, `ms-`, `me-`) + an RTL-aware layout. Toggle font stack per locale (Arabic web font).
- Language switcher persists choice (localStorage + `?lang=`), and the choice flows into every API call (`?lang=` or `Accept-Language`).

---

## 8. PWA + Push Notifications

**PWA**
- `vite-plugin-pwa` (Workbox under the hood). Generate `manifest.webmanifest` (name, icons, theme color, `display: standalone`).
- Service worker caching: precache app shell; runtime cache for images and `GET /menu` (stale-while-revalidate). App works offline for browsing the menu.
- "Add to Home Screen" prompt handling.

**Push notifications (Web Push)**
- **VAPID** key pair (public key in frontend, private key in backend env).
- Flow: user opts in → browser `PushManager.subscribe` → POST subscription to `/push/subscribe` → store in `push_subscriptions`.
- Backend sends via an Asynq job using a Web Push library (Go: `SherClockHolmes/webpush-go`) on order-status changes ("Your order is out for delivery 🚗").
- iOS: works for installed PWAs (iOS 16.4+). Document the install-first requirement for iOS users.

---

## 9. Orders & Real-time Status

- **Cart** lives client-side (React Query + a cart store); validated server-side at checkout (re-price everything from DB — never trust client prices).
- **Checkout** = the idempotent `POST /orders` described in §3.3.
- **Customer order tracking**: `GET /orders/:code` + **live updates** via SSE (`GET /orders/:code/stream`) or WebSocket. Backend `realtime` hub subscribes to a Redis Pub/Sub channel per order; when staff advance the status, the worker publishes → all connected clients update instantly. SSE is simpler and enough here.
- **Admin/kitchen order board**: list + filter by status, advance status (writes `order_status_history`, publishes update, enqueues customer push).

---

## 10. Payments — Square (Canada) + Cash on Delivery, each toggleable

**Abstraction:** define a `PaymentProvider` interface in `internal/payment` so providers are pluggable. Each provider has an **on/off flag in `settings`**, so an admin can enable/disable a method without a deploy. The checkout only offers methods whose flag is `true` and whose keys are configured.

```go
type PaymentProvider interface {
    Key() string                                                   // "square", "cod"
    Enabled(cfg Settings) bool                                     // settings-driven on/off
    CreatePayment(ctx, order) (clientToken string, ref string, err error)
    HandleWebhook(ctx, payload, signature string) (PaymentEvent, error)
    Refund(ctx, txnRef string, amountCents int64) error
}
```

### 10.1 Square (online card) — confirmed available in Canada

- **Availability & currency:** Square's payment APIs operate in Canada, processing in **CAD** and depositing to a Canadian bank account in ~1–2 business days. No separate merchant account or gateway needed.
- **Integration shape — Web Payments SDK + Payments API (recommended):**
  1. **Frontend** loads the Square Web Payments SDK, renders the card field (also enables **Apple Pay / Google Pay**), and tokenizes the card into a single-use `token` (PAN never touches your server → keeps you out of heavy PCI scope).
  2. **Backend** receives the `token`, calls Square's **Payments API** (`POST /v2/payments`) with amount in CAD cents, an **idempotency key**, and the order reference. Set the `Square-Version` header (e.g. `2026-05-20`).
  3. On success → mark order `paid`, advance to `confirmed`, enqueue notifications.
- **Alternative — Checkout API (hosted):** generate a Square-hosted payment link and redirect; lower effort, less UI control. Keep as a fallback path behind the same interface.
- **Webhooks:** subscribe to Square payment events; **verify the signature**, then process **idempotently** in an Asynq `critical` job (handles the case where the SDK callback is lost but the charge succeeded). Update `payment_transactions` + order status from the webhook as the source of truth — never trust a client-side "success".
- **Refunds:** Square Refunds API via `Refund(...)`, triggered from the admin order screen.
- **Keys (all in `settings`, encrypted, sandbox + production sets):** `square_enabled` (bool), `square_environment` (`sandbox`/`production`), `square_application_id`, `square_location_id`, `square_access_token`, `square_webhook_signature_key`. Toggling `square_enabled` off instantly removes the card option from checkout.
- **Sandbox first:** build and test entirely against Square Sandbox; flipping `square_environment` to `production` with live keys is the only switch needed to go live.

### 10.2 Cash on Delivery (COD)

- A no-op provider gated by `cod_enabled` (bool). When chosen: order goes straight to `confirmed` with `payment_status='cod_pending'`, settled/marked `paid` by staff on delivery.
- Optional guardrails (in `settings`): max COD order value, restrict COD to certain delivery zones.

### 10.3 Method selection logic (checkout)

```
available_methods = [p for p in providers if p.Enabled(settings) and p.KeysConfigured()]
// e.g. settings: square_enabled=true, cod_enabled=true  → customer sees [Card (Square), Cash on Delivery]
//      square_enabled=false                              → customer sees [Cash on Delivery] only
```

The interface + settings flags mean adding a future provider (e.g. another gateway) is just a new `PaymentProvider` implementation and one settings flag — no checkout rewrite.

---

## 11. Coupons

- Validate at checkout: active, within date window, under usage/per-user limits, meets `min_order_cents`.
- Types: percent (with `max_discount_cents` cap), fixed, free delivery.
- Atomic `used_count` increment inside the order transaction to prevent over-redemption under concurrency (row lock or `UPDATE ... WHERE used_count < usage_limit`).
- Admin CRUD + usage analytics.

---

## 12. Banners / Offers

- `banners` CRUD with image, link, schedule (`starts_at/ends_at`), sort order, active flag.
- Public endpoint returns only currently-active banners (cached, TTL 5 m, invalidated on edit).
- Frontend renders a hero carousel; translatable titles.

---

## 13. User & Order Management (Admin)

- **RBAC**: roles `customer | staff | kitchen | admin | super_admin`; middleware checks permissions per route. (Matches your prior RBAC work.)
- **Users**: list/search/filter, view order history, change role, suspend.
- **Orders**: full board, status transitions, refunds, notes, manual order creation.
- **Catalog**: products, variants, modifiers, categories, availability toggles, pre-order config.
- **Settings**: languages, Meta keys, payment keys, tax %, store hours, delivery zones map editor, cache flush.

---

## 14. Meta Integration (key-based, easy to wire)

- Settings page fields: **Meta Pixel ID**, **Conversions API access token**, **dataset/pixel id**, optional test event code.
- **Frontend**: inject Pixel script only if a Pixel ID is set; fire standard events (`ViewContent`, `AddToCart`, `InitiateCheckout`, `Purchase`).
- **Backend (Conversions API)**: on `Purchase`/`InitiateCheckout`, an Asynq job sends a server-side event to Meta with the configured token (better attribution, hashed PII). All driven by the keys in `settings` — drop in keys, it activates; no code change.
- Same settings-driven pattern leaves room for WhatsApp Business / Messenger later.

---

## 15. Frontend Architecture

```
frontend/
├── src/
│   ├── app/            # router, providers (QueryClient, i18n, theme, auth)
│   ├── api/            # axios/fetch client + typed endpoints + query keys
│   ├── features/       # menu, cart, checkout, orders, account, auth
│   ├── components/     # shared UI (Tailwind), RTL-aware
│   ├── hooks/          # useGeolocation, useCart, usePush, etc.
│   ├── i18n/           # i18next config + loaders
│   ├── pwa/            # SW registration, push subscribe
│   ├── stores/         # cart store (zustand or context)
│   └── types/          # shared TS types (mirror API DTOs)
├── vite.config.ts      # + vite-plugin-pwa
└── tailwind.config.ts  # v3, RTL plugin, brand tokens
```

- **TanStack Query** for all server state: query keys mirror cache namespaces; `staleTime` tuned per resource (menu long, order status short / use SSE). Mutations invalidate the right query keys (cart, orders).
- **Design**: modern, image-forward (your menu photos are strong) — warm/amber palette matching the menu, large product cards, sticky cart, smooth micro-interactions. Keep cards a fixed, font-scale-safe height (you've hit the RN large-font card-clipping bug before — same discipline here with `line-clamp` + min-heights).
- **Admin** is a separate Vite app (or a `/admin` route guarded by RBAC) sharing the API client + types.

---

## 16. Phased Build Order (hand each phase to the assistant in turn)

1. **Foundation** — repo, Docker Compose (MySQL, Redis, Meilisearch), Go skeleton (config, logger, response envelope, Echo + middleware), migrations tooling, `.env.example`. Health check endpoint.
2. **Auth & users** — JWT auth, RBAC middleware, user CRUD, phone (E.164) + Canada address validation.
3. **Catalog** — categories, products (unit + weight modes), variants, modifiers, translations, admin CRUD, public menu endpoint + cache-aside + invalidation.
4. **Search** — Meilisearch wiring, index sync jobs, `/search` endpoint.
5. **i18n** — languages table, translation store, locale bundles, admin language management, frontend i18next + RTL.
6. **Cart & orders** — checkout (idempotent, server-priced), order model, status history, COD path.
7. **Geo & delivery zones** — Haversine + polygon matching, global + per-product zones, admin map editor, fee calc.
8. **Payments** — `PaymentProvider` interface, **Square (Web Payments SDK + Payments API)** in sandbox then production, settings on/off toggles for Square + COD, idempotent webhooks, refunds.
9. **Coupons & banners** — CRUD, validation, public endpoints.
10. **Real-time + push** — SSE/Redis pub-sub order tracking, Web Push subscriptions + notification jobs.
11. **PWA** — manifest, service worker, offline menu, install prompt.
12. **Meta integration** — settings-driven Pixel + Conversions API.
13. **Hardening** — rate limits, idempotency everywhere, structured logging on every order/payment transition, graceful shutdown, load test, cache-flush admin tooling.

**Acceptance gates per phase:** migrations run clean, endpoints documented (OpenAPI), unit tests on services, and a smoke test of the happy path.

---

## 17. Non-functional checklist

- Integer cents everywhere; `CAD` currency; tax configurable per province.
- Structured logs with request IDs; every order/payment state change logged (zero silent failures).
- Idempotency keys on orders + payment webhooks.
- Soft deletes + `order_status_history` audit trail.
- Secrets in env / encrypted settings, never in the repo.
- `docker-compose` for local; same images for prod with N API replicas + dedicated worker process.
- Backups for MySQL; Meilisearch is rebuildable from MySQL (full reindex job).
