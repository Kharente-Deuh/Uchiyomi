# 9. Vuetify as the UI component framework

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi is a mobile-first, installable PWA that needs a broad, accessible
component set (app bars, lists, dialogs, menus, theming, dark mode) so effort goes
into reader/library UX rather than building primitives. The scaffold is based on
vitesse-nuxt, which ships UnoCSS.

## Decision

Use **Vuetify 4** via **`vuetify-nuxt-module`** as the UI layer. Remove UnoCSS and
`@nuxtjs/color-mode` from the vitesse base; Vuetify's theme system owns light/dark
(`defaultTheme: 'system'` via SSR client hints). Vuetify options live in
`vuetify.config.ts`.

Icons use **Font Awesome 6 via Iconify**, wired as a custom Vuetify icon set
(`defaultSet: 'custom'`, configured in the `vuetify-icons` plugin through the
module's `vuetify:configuration` hook). Icon data is bundled offline from
`@iconify-json/fa6-*` and rendered synchronously via `@iconify/vue`'s `getIcon`
(SSR-safe). This deliberately avoids UnoCSS — the easiest Iconify path
(`presetIcons`) would reintroduce it, which this ADR rules out — so colours stay
owned by the theme (`currentColor`). MDI (`@mdi/font`) is not used.

## Alternatives considered

- **UnoCSS (vitesse default) + headless components** — maximum control and tiny
  CSS, but means building/maintaining a component layer ourselves.
- **Nuxt UI / PrimeVue / Quafu** — viable, but Vuetify offers the most complete
  Material component set and first-class theming for the dashboard-style admin and
  reader surfaces planned here.

## Consequences

- `vuetify-nuxt-module` is currently **`1.0.0-rc.1`** (a release candidate, not
  yet a stable 1.0.0). It is the Nuxt 4 + Vuetify 4 compatible line; the last
  stable line (`0.19.x`) targets Nuxt 3 / Vuetify 3. We accept tracking the RC and
  will move to stable `1.0.0` when released.
- `sass-embedded` is required for Vuetify theming/build.
- Keeping a single styling system (Vuetify) avoids UnoCSS/Vuetify duplication.
- Component-heavy bundle vs a utility-CSS approach; acceptable for an
  authenticated app (not a public marketing site).
