# 4. Shared Suwayomi engine with a per-user PostgreSQL overlay

- Status: Accepted
- Date: 2026-06-20

## Context

Suwayomi is fundamentally single-user: one shared library and one shared read
state; true multi-user accounts are a long-standing, unshipped feature. Uchiyomi
needs real per-user libraries and reading history. The deployment target is a NAS
(modest RAM), shared by a household.

## Decision

Run **one shared Suwayomi instance** as a pure source/fetch/catalogue/page engine,
and own the **per-user layer in PostgreSQL** (Uchiyomi's "app" database):
subscriptions ("my library"), categories, and reading progress, referencing
Suwayomi IDs.

Two rules:
1. Suwayomi `inLibrary = true` means "at least one user follows this title" (so its
   updater/downloader runs). "Who follows it" lives in the `subscription` table.
   Downloads are shared on disk → de-duplicated.
2. Suwayomi's global `isRead` flag is never used for per-user semantics; progress
   lives only in `reading_progress`.

## Alternatives considered

- **One Suwayomi instance per user** — rejected for a NAS: N JVMs (heavy RAM),
  duplicated downloads (no de-dup), N update jobs. Only viable for very few users
  with ample RAM.

## Consequences

- Elegant on a NAS: shared, de-duplicated downloads; isolated per-user state.
- Uchiyomi does not reuse Suwayomi's own library/progress; it reimplements that
  layer — which is the part worth owning.
- Per-user catalogue access control (e.g. NSFW filtering) must be enforced in the
  overlay, since Suwayomi's catalogue is global.
