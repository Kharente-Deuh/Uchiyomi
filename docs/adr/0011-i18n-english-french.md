# 11. Internationalization with English and French

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi targets a household/self-hoster audience that is not English-only. UI
strings must be translatable from the start so localization is not retrofitted.
Initial languages are English and French.

## Decision

Use **`@nuxtjs/i18n`** with **English (`en`) and French (`fr`)**. Locale messages
are lazy-loaded JSON in `i18n/locales/{en,fr}.json`; vue-i18n runs in Composition
mode (`legacy: false`) with `en` as default and fallback. Strategy is
`no_prefix` (no locale in the URL) with cookie-based browser-language detection.

## Alternatives considered

- **English-only now, i18n later** — rejected: retrofitting translations across an
  existing UI is costly and error-prone.
- **`strategy: 'prefix_except_default'`** (locale in URL) — better for public,
  SEO-driven sites; unnecessary for an authenticated self-hosted app, so the
  simpler `no_prefix` + cookie is preferred.

## Consequences

- All user-facing strings go through i18n keys; no hard-coded copy in components.
- `en` and `fr` locale files must stay key-for-key in sync; a Vitest test asserts
  key parity to catch drift.
- Adding a language = a new locale file + a `locales` entry; no structural change.
- Vuetify's own component locale messages can be wired to the active locale later
  if needed.
