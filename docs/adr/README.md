# Architecture Decision Records

This directory records significant architectural decisions for Uchiyomi, one file
per decision, using the [Nygard format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

ADRs are **immutable**: when a decision changes, add a new ADR that supersedes the
old one rather than editing it.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](./0001-use-suwayomi-as-headless-engine.md) | Use Suwayomi-Server as a headless source engine | Accepted |
| [0002](./0002-downloaded-first-reading-model.md) | Downloaded-first reading model | Accepted |
| [0003](./0003-single-unified-app-no-komga.md) | Single unified app; do not use Komga as the user-facing layer | Accepted |
| [0004](./0004-shared-engine-per-user-overlay.md) | Shared Suwayomi engine with a per-user PostgreSQL overlay | Accepted |
| [0005](./0005-nuxt-nitro-monolith.md) | Backend as a Nuxt/Nitro monolith, not a separate NestJS service | Accepted |
| [0006](./0006-auth-local-and-sso.md) | Local accounts + SSO with revocable sessions and pragmatic RBAC | Accepted |
| [0007](./0007-license-agpl-3.md) | License: AGPL-3.0-or-later | Accepted |
| [0008](./0008-prisma-overlay-data-access.md) | Prisma as the data-access layer for the PostgreSQL overlay | Accepted |
| [0009](./0009-vuetify-ui-layer.md) | Vuetify as the UI component framework | Accepted |
| [0010](./0010-frontend-scaffold-and-tooling.md) | Frontend scaffold and tooling baseline | Accepted |
| [0011](./0011-i18n-english-french.md) | Internationalization with English and French | Accepted |
| [0013](./0013-application-structure.md) | Application structure: atomic-design frontend, DDD server | Accepted |
