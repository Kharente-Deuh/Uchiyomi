# Contributing to Uchiyomi

Thanks for your interest in contributing!

## Language

**English everywhere in the repository** — code, identifiers, comments, commit
messages, docs, issues, and PR titles/descriptions.

## Prerequisites

- **Node.js 26** (see `.nvmrc`)
- **pnpm 11** (`corepack enable` will pick up the pinned version)

## Setup

```bash
pnpm install
pnpm dev
```

## Useful scripts

| Script            | What it does                                  |
| ----------------- | --------------------------------------------- |
| `pnpm dev`        | Start the dev server                          |
| `pnpm build`      | Production build                              |
| `pnpm typecheck`  | `vue-tsc` type checking (strict)              |
| `pnpm lint`       | ESLint (`@antfu/eslint-config`)               |
| `pnpm lint:fix`   | ESLint with autofix                           |
| `pnpm knip`       | Find unused files, deps and exports           |
| `pnpm test`       | Run the Vitest suite once                     |
| `pnpm test:watch` | Vitest in watch mode                          |
| `pnpm up`         | Interactive dependency upgrades (taze)        |

## Commits

We use [Conventional Commits](https://www.conventionalcommits.org/). Types:
`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `ci`, `build`,
`style`, `revert`. Breaking changes use `!` and/or a `BREAKING CHANGE:` footer.
`commitlint` enforces this on a `commit-msg` git hook.

## Before you push

A `pre-push` hook runs `typecheck`, `lint`, `knip` and `test`. A change is
"done" only when all of these pass and behavior changes ship with tests.

## Architecture

Read [`CLAUDE.md`](./CLAUDE.md) and the ADRs in [`docs/adr/`](./docs/adr/).
Cross-cutting or architectural decisions should come with a new ADR.

## License

By contributing you agree your contributions are licensed under
**AGPL-3.0-or-later**.
