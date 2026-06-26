<p align="center">
  <img src="app/assets/images/logo-horizontal-subtitle.webp" style="width: 60%">
</p>
<p align="center">
  <span>Opiniated and self-hosted manga & webtoon server with a modern, installable PWA. Based on <a href="https://github.com/Suwayomi/Suwayomi-Server">Suwayomi</a></span>
</p>



[![CI](https://github.com/Kharente-Deuh/Uchiyomi/actions/workflows/ci.yml/badge.svg)](https://github.com/Kharente-Deuh/Uchiyomi/actions/workflows/ci.yml)
[![coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/Kharente-Deuh/2cc6c60da922faeb492b67b2681f6e9c/raw/uchiyomi-coverage.json)](https://github.com/Kharente-Deuh/Uchiyomi/actions/workflows/ci.yml)
[![Latest stable release](https://img.shields.io/github/v/release/Kharente-Deuh/Uchiyomi?label=stable)](https://github.com/Kharente-Deuh/Uchiyomi/releases/latest)
[![Latest release](https://img.shields.io/github/v/release/Kharente-Deuh/Uchiyomi?include_prereleases&label=latest)](https://github.com/Kharente-Deuh/Uchiyomi/releases)
---

Uchiyomi runs [Suwayomi-Server](https://github.com/Suwayomi/Suwayomi-Server)
headless as its source/download engine (Tachiyomi/Mihon extensions) and adds the
layer Suwayomi lacks: **multi-user accounts, SSO, per-user libraries, and
per-user reading history**.

> **Status:** early development. This is the application scaffold — see
> [`CLAUDE.md`](./CLAUDE.md) and the ADRs in [`docs/adr/`](./docs/adr/) for the
> vision and architecture.

See [**Choices vs Suwayomi**](./docs/suwayomi-choices.md) for where Uchiyomi
wraps, overrides, or diverges from Suwayomi.

## Tech stack

- **Nuxt 4 (Vue 3) + Nitro** — single deployable
- **Vuetify** — UI component framework
- **PWA** via `@vite-pwa/nuxt`
- **i18n** (English + French) via `@nuxtjs/i18n`
- **TypeScript** (strict)
- **PostgreSQL** overlay for the per-user layer (planned)
- Tooling: `@antfu/eslint-config`, Vitest, knip, taze, husky, commitlint

## Development

Requires **Node 26** and **pnpm 11** (`corepack enable`).

```bash
pnpm install
pnpm dev
```

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for scripts and conventions.

## License

[AGPL-3.0-or-later](./LICENSE). See [ADR-0007](./docs/adr/0007-license-agpl-3.md).
