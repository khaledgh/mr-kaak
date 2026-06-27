# Restaurant PWA — Frontend

React + Vite + TypeScript + TailwindCSS v3 + TanStack Query + i18next (RTL-aware) + vite-plugin-pwa.

> **Status:** Customer PWA — core flows complete (menu, cart, checkout, live
> order tracking, auth, language/RTL switch, installable PWA). The admin SPA is
> a separate follow-up (shares the same API client + types).

## Run

```bash
npm install
npm run dev      # http://localhost:5173 — proxies /api to the Go backend on :8080
npm run build    # type-check (tsc) + production build + service worker
npm run preview  # serve the production build
```

Start the backend (`go run ./cmd/api` in `../backend`) first so `/api` resolves.

## Structure

```
src/
  api/         axios client (token + locale interceptors, refresh) + typed endpoints + query keys + DTO types
  app/         router (route guards)
  components/  Layout, Header, LanguageSwitcher, CartDrawer mount
  features/
    menu/      MenuPage (banners + search + categories), ProductCard
    cart/      CartDrawer (slide-over, qty edit)
    checkout/  CheckoutPage (fulfillment, address, payment, coupon, idempotent submit)
    orders/    OrdersPage (history), OrderTrackPage (SSE live status stepper)
    auth/      LoginPage, RegisterPage
  i18n/        i18next config + en/ar/fr bundles; applyDir() sets <html dir> for RTL
  stores/      cart + auth (zustand, persisted to localStorage)
  lib/         money formatting (Intl)
```

## Key behaviours

- **Server-priced checkout:** the cart shows display estimates; the server
  re-prices every line at checkout. An `Idempotency-Key` (UUID) is sent so a
  retried submit never double-orders.
- **Live tracking:** `OrderTrackPage` opens an SSE connection
  (`/orders/:code/stream?access_token=…`) for instant status updates, falling
  back to the fetched status if the stream drops.
- **i18n + RTL:** language switcher is driven by the server's active-languages
  list; choosing Arabic flips `<html dir="rtl">` and swaps the font stack.
- **PWA:** installable; app shell precached; menu (stale-while-revalidate) and
  images (cache-first) cached for offline browsing.

## TODO (follow-ups)

- Admin SPA (catalog/orders/settings/zone-map editor).
- Square Web Payments SDK card field on checkout when `square` is enabled.
- Web Push opt-in UI (`/push/subscribe`) once VAPID keys are set on the backend.
- Replace placeholder PWA icons (`public/icon-192.png`, `icon-512.png`).
