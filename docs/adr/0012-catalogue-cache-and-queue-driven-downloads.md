# 12. Catalogue cache, admin-managed extensions, and queue-driven downloads

- Status: Accepted
- Date: 2026-06-20

## Context

The product needs a "latest releases" home feed (series sorted by last update, with
per-chapter read state, hiatus/resume badges), controlled per-user access to sources
(originally framed as private per-user sources; see the revised Decision below), and
controlled, non-blocking downloads. Suwayomi exposes this data only through its API
and paces downloads with its own queue. Querying it live per feed render, and
relying on its auto-download, did not meet the performance and control goals.

## Decision

Three overlay decisions, all keeping Suwayomi as the source of truth (ADR-0001):

1. **Cached catalogue.** Replicate the series/chapter metadata the feed needs into
   PostgreSQL (`series`, `chapter`), for *followed* series only. Synced by a cron and
   by Suwayomi events. We cache, we do not reparse sources.
2. **Admin-managed extensions (revised 2026-06-22, M4.1a).** Extensions are
   installed once, globally, in Suwayomi by an admin (`canManageExtensions`);
   every user sees the installed set and browses within it. Per-user activation
   (`user_extension_activation`) and refcount-driven auto-uninstall are **dropped**.
   NSFW visibility is a per-user **hybrid gate**: `allowNsfw` (admin-granted right)
   AND `showNsfw` (self-serve preference). Extension health stays on `extension`
   plus an `extension_error_log` history for admins.
3. **Queue-driven downloads (amends ADR-0004).** All downloads — initial backfill and
   new chapters — flow through an overlay `download_job` queue paced by a single
   worker, so throttling is global. Suwayomi's updater still detects new chapters, but
   its auto-download is disabled. New-chapter jobs outrank backfill.

`series.status` adds a derived `HIATUS` (no chapter for > 1 month, not completed),
recomputed weekly; the "reprise" badge is event-driven via `series.resumed_at`
(set on a new chapter after a > 1-month gap, shown for 7 days). Extension health is
tracked on `extension` plus an `extension_error_log` history for admins.

### Suwayomi settings reconciliation (added 2026-06-24)

A fresh Suwayomi-Server ships with **no extension repositories**
(`settings.extensionRepos = []`); with none, `fetchExtensions` pulls nothing and
`listAvailable()` returns `[]`. **Uchiyomi is the single source of truth for the
Suwayomi settings it depends on.** A boot-time **reconciler** (a Nitro startup plugin)
reads the relevant settings and brings them into line via `setSettings`, with a
per-setting policy:

- **`ENFORCE`** — always rewritten to the desired value (invariants). Notably
  `autoDownloadNewChapters = false`, defence in depth over decision n°3.
- **`SEED_IF_EMPTY`** — a default written only when empty; admin edits persist. Notably
  `extensionRepos` seeded with the Keiyoushi community repo so the catalogue works out
  of the box.
- **`FREE`** — never touched by the reconciler; admin-managed via the settings panel
  (M6.2).

Repos are therefore provisioned by Uchiyomi, **not** by a Suwayomi-side
`EXTENSION_REPOS` Docker env var (which would clobber in-app edits on container
restart). The desired-settings list is code-defined and seeds the curated allowlist the
M6.2 admin panel will extend. Spec of record:
`docs/superpowers/specs/2026-06-24-suwayomi-settings-reconciler-design.md`.

## Consequences

- The feed and read/unread colouring become SQL joins instead of N Suwayomi calls.
- We own catalogue-sync, download-pacing, and extension-health code that Suwayomi
  partly provided — the trade for global control and feed performance.
- ADR-0004's "Suwayomi's downloader runs" is superseded: Suwayomi downloads only when
  our queue tells it to.
- Spec of record: `docs/superpowers/specs/2026-06-20-overlay-data-model-design.md`.
