# 10. Frontend scaffold and tooling baseline

- Status: Accepted
- Date: 2026-06-20

## Context

The app needed a starting Nuxt 4 project plus the engineering tooling the project
mandates (linting, commit hygiene, dependency hygiene, releases). Rather than
assemble from scratch, we start from a maintained template.

## Decision

Scaffold from **`antfu/vitesse-nuxt`** (Nuxt 4, Vue 3, Nitro, PWA, Pinia, VueUse),
stripped of demo content, and standardize on:

- **Package manager:** pnpm 11 with the **pnpm catalog** (`pnpm-workspace.yaml`);
  **Node 26** (`.nvmrc`, `engines`; even-numbered → Active LTS from Oct 2026,
  and a version Nuxt's `engines` field lists, unlike Node 25).
- **Lint/format:** `@antfu/eslint-config` (single tool; UnoCSS preset removed),
  integrated with `@nuxt/eslint`.
- **Commits:** Conventional Commits enforced by **commitlint**, via **husky**
  hooks — `commit-msg` (commitlint), `pre-commit` (lint-staged), `pre-push`
  (`typecheck → lint → knip → test`).
- **Dependency hygiene:** **knip** (unused files/deps/exports) and **taze**
  (`pnpm up` = interactive upgrades).
- **Releases/CI:** **release-please** (SemVer from commit history) and
  **Renovate** (automated dependency-update PRs); GitHub Actions for
  `typecheck → lint → knip → test → build`.

## Alternatives considered

- **`nuxi init` bare template** — less batteries-included; we would re-add PWA,
  Pinia, VueUse, ESLint integration manually.
- **lefthook** instead of husky — fine, but husky was the requested tool; either
  satisfies the hook requirement.
- **Dependabot** instead of Renovate — Renovate offers grouping and broader config.

## Consequences

- `@antfu/eslint-config` supersedes the `@nuxt/eslint`-only convention noted in
  earlier docs; `eslint.config.js` is the place to add project-specific options.
- **Supply-chain note:** vitesse set pnpm `trustPolicy: no-downgrade`, which
  blocked install on legacy *transitive* deps (workbox/babel under
  `@vite-pwa/nuxt`) lacking npm provenance. It was removed to unblock install — a
  deliberate relaxation to revisit (re-enable with `trustPolicyExclude`, or adopt
  `minimumReleaseAge`).
- `vuetify-nuxt-module` replaces vitesse's UnoCSS (see ADR-0009); data access uses
  Prisma (see ADR-0008); localization uses `@nuxtjs/i18n` (see ADR-0011).
