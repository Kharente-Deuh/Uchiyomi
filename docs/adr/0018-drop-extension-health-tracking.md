# 18. Drop extension-health tracking

- Status: Accepted
- Date: 2026-06-28

## Context

ADR-0012 (revised M4.1a) and ADR-0017 introduced extension-health tracking as a
feature we own on top of Suwayomi: an `ExtensionHealth` enum (`OK` / `ERROR`),
four columns on the `extension` overlay row (`health`, `consecutive_failures`,
`last_error_at`, `last_error_message`), and an `extension_error_log` history table.
Install/update use cases recorded success/failure into it, the list endpoint
annotated each item with `isHealthy`, and `GET /extensions/:pkgName` exposed the
full `ExtensionHealthDto` (counter + error log).

In practice the feature was never surfaced. It was plumbed end to end — DB →
domain → use cases → route → frontend store/composable — but **no UI ever
consumed it**: `isHealthy`, `ExtensionHealthDto`, the `extensions.health.*` /
`extensions.filters.healthy` i18n keys, and the planned health badge/filter were
all dead. It carried real cost: an extra table and four columns, a transaction on
every install/update, a use case, repository methods, a presenter, DTOs, and the
frontend wiring to keep them flowing.

## Decision

**Remove extension-health tracking entirely.** Specifically:

- Drop the `ExtensionHealth` enum, the four health columns on `extension`, and the
  `extension_error_log` table (new migration `drop_extension_health`).
- Remove the `GetExtensionHealthUseCase`, the overlay health methods
  (`listHealthByPkgNames` / `findHealth` / `recordFailure` / `recordSuccess` /
  `listErrorLog`), the `toHealthDto` presenter, and `ExtensionHealthDto`.
- Drop `isHealthy` from `ExtensionDto`; collapse `ListedExtension` /
  `toListedExtension` back into the plain `ExtensionModel` returned by the list /
  install / uninstall / update use cases. The install/update use cases no longer
  wrap the Suwayomi call to record outcomes — errors simply propagate.
- `GET /extensions/:pkgName` now returns `{ extension }` only.
- Remove the corresponding frontend state, the `extensions.health.*` /
  `extensions.filters.healthy` i18n keys, and the dead `isHealthy` filter field.

Suwayomi remains the source of truth for extension state (installed / has-update /
nsfw); the overlay `extension` row keeps only its install trace
(`installed_by_user_id`, `installed_at`).

## Alternatives considered

- **Keep it and finally build the UI.** Rejected: there is no concrete product
  need for a per-extension health indicator, and the recording cost is paid on
  every install/update regardless. If the need returns, Suwayomi failures already
  surface through `SuwayomiError`, and a fresh, UI-driven design will fit better
  than the speculative model built here.
- **Keep the columns, drop only the unused API/UI surface.** Rejected: dead
  columns and an unused history table are still schema and write-path cost; a clean
  removal is simpler to reason about.

## Consequences

- Amends ADR-0012 and ADR-0017: the `extension` overlay row no longer carries
  health fields, there is no `extension_error_log` table, `ExtensionDto` has no
  `isHealthy`, and `GET /extensions/:pkgName` no longer returns an
  `ExtensionHealthDto`. The three-lever source-visibility rule, dormancy semantics,
  and per-source `isEnabled` from ADR-0017 are unaffected.
- The M4.1b browse UI drops the "extension-health indicator" open question.
- Should health tracking be revived, it will need a new ADR and a UI-first design.
