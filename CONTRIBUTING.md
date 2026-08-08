# Contributing

Thanks for taking the time to contribute to Uchiyomi.

Everything in this project — code, comments, documentation, commit messages and
UI strings — is written in English.

## Getting started

Pre-requisites and the initial setup are listed in the [README](./README.md).
In short:

```bash
pnpm --dir web install
lefthook install
```

`lefthook install` is not optional: the git hooks add the license header, check
commit messages and run lint/build/test before a push.

## Running the project

```bash
docker compose up -d uchiyomi-db          # Postgres on :5432
cd api && go run ./cmd/uichiyomiserver    # API on :3000
cd web && pnpm dev                        # SPA on :3001
```

The Go binary is configured through environment variables only — no config file,
no flags besides `-healthcheck`. `DB_HOST`, `DB_USER`, `DB_PWD` and `DB_NAME`
are required; see `api/cmd/uichiyomiserver/config.go` for the full list.

In dev, Nuxt proxies `/api` to `:3000`. In production the SPA is embedded into
the binary through the `webui` build tag (`WITH_WEB=on`), so a single container
serves both. To exercise that path:

```bash
docker compose up --build
```

## Scripts

### api/

```bash
go build ./...
go test ./...
go test ./pkg/core/auth -run TestServiceLogin   # single test
golangci-lint run ./...
```

### web/

```bash
pnpm build        # nuxt generate → .output/public
pnpm test         # vitest run
pnpm lint         # pnpm lint:fix to autofix
pnpm typecheck
pnpm knip         # dead code detection, must stay green
```

## Conventions

### Code

The codebase carries no explanatory comments. The only ones allowed are:

- the `SPDX-License-Identifier: AGPL-3.0-or-later` header every source file
  starts with (added automatically by the pre-commit hook),
- `TODO` markers,
- linter bypasses (`//nolint:…`, `eslint-disable…`) and compiler directives
  (`//go:build`, `//go:embed`, `//go:generate`).

Prefer a precise name and a clear signature over a comment.

### Linting

Both sides are linted in CI-equivalent hooks and the rules are opinionated —
`gochecknoglobals`, `lll`, `nlreturn`, `dupl` and `govet.fieldalignment` on the
Go side, `@antfu/eslint-config` plus knip on the web side. Match the existing
style rather than adding bypasses; a `//nolint` needs a reason.

### Tests

Go tests live in `<name>_test.go` in the `_test` package; use
`<name>_internal_test.go` only when the test genuinely needs unexported symbols.
Repository tests use `pkg/repository/pgtest` (sqlmock with the real Postgres
dialect), so no container is required to run `go test ./...`.

On the web side, logic is tested where it is pure — routing rules in
`app/utils/guards/`, form helpers in `app/utils/forms/` — rather than through
the Nuxt middleware wrappers.

### Commits

One short line, [Conventional Commits](https://www.conventionalcommits.org),
enforced by the `commit-msg` hook:

```
<type>[(scope)][!]: <description>
```

Allowed types: `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`,
`refactor`, `revert`, `style`, `test`.

```
feat(web): add vertical reader
fix(api)!: rename the chapterId field
```

If your branch carries a ticket number, add it as a final line of the message.

## Pull requests

1. Branch off `master`.
2. Keep the change focused — one coherent unit of work per PR.
3. Make sure the pre-push hook passes for the side you touched (lint, build,
   test, and for `web/` also typecheck and knip). `--no-verify` exists but a red
   hook is a red PR.
4. Describe what changed and why; link the issue if there is one.

## License

By contributing you agree that your contributions are licensed under
[AGPL-3.0-or-later](./LICENSE).
