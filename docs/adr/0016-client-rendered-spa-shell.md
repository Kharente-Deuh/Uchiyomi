# 16. Client-rendered SPA shell (`ssr: false`)

- Status: Accepted
- Date: 2026-06-21

## Context

Uchiyomi is fully authenticated (no public pages, no SEO), an installable PWA
with a service worker, and self-hosted on potentially modest hardware alongside
Suwayomi, PostgreSQL, and the in-process download queue (ADR-0012/0014). The
shell milestone (M3.2) forced a rendering decision: server-side render the Vue
app, or render it on the client.

## Decision

Set `ssr: false` and keep `nitro.prerender.routes: ['/']`, so Nitro serves a
**prerendered static SPA shell** that the service worker caches; the client
hydrates and loads the user via `GET /api/auth/me`. This is a rendering choice
only — the Nitro server runtime is unchanged: all `/api/*` routes, the Suwayomi
client, Prisma, and the download queue keep running server-side.

Navigation is an **adaptive shell**: a bottom navigation bar on mobile and a
navigation rail on desktop, built from pure Organism components fed by a global
destination registry (`useNavigation`). A client boot plugin hydrates the
identity store before a global route middleware gates routes.

## Rationale

- **SEO** — the usual reason to SSR — does not apply (everything is behind login).
- **First paint:** SSR would call `/api/auth/me` (a DB hit) before responding on
  each full navigation; the prerendered shell paints instantly instead.
- **PWA/offline:** the service worker already serves the cached `/` shell, so SSR
  of authenticated content is wasted on installed launches — client rendering is
  the canonical PWA "app shell" model.
- **Server load:** keeps Nitro a lean API + static server on small hardware.
- **Auth correctness:** no SSR cookie forwarding and no hydration-mismatch flashes
  of wrong auth state.

## Security note (threat model)

With `ssr: false` the route middleware runs **client-side only**. The auth guard
is therefore **UX routing, not a security boundary** — a user cannot reach
protected data without a valid server session, because every `/api/*` route
enforces the session server-side (M2.2, ADR-0006). The client guard only avoids
rendering empty authenticated shells to signed-out visitors.

## Alternatives considered

- **Keep SSR (Nuxt default):** better first paint for *public* content on capable
  servers — neither condition holds here; adds per-request render cost and
  SSR-auth complexity.
- **Client-only with no prerender:** loses the instant static shell; rejected in
  favour of `ssr: false` + prerendered `/`.

## Consequences

- The shell and all pages render on the client; route middleware is UX-only.
- Server stays a pure API + static host; the download queue is unaffected.
- The PWA manifest `orientation` lock is wired later (M6.1); the in-app portrait
  lock ships now (M3.2) via `useOrientationLock`.
