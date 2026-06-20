# 12. Catalogue cache, per-user extension activation, and queue-driven downloads

- Status: Accepted
- Date: 2026-06-20

## Context

The product needs a "latest releases" home feed (series sorted by last update, with
per-chapter read state, hiatus/resume badges), private per-user sources, and
controlled, non-blocking downloads. Suwayomi exposes this data only through its API
and paces downloads with its own queue. Querying it live per feed render, and
relying on its auto-download, did not meet the performance and control goals.

## Decision

Three overlay decisions, all keeping Suwayomi as the source of truth (ADR-0001):

1. **Cached catalogue.** Replicate the series/chapter metadata the feed needs into
   PostgreSQL (`series`, `chapter`), for *followed* series only. Synced by a cron and
   by Suwayomi events. We cache, we do not reparse sources.
2. **Per-user extension activation.** Extensions stay installed once, globally, in
   Suwayomi; a per-user `user_extension_activation` overlay makes "my sources"
   private. The first activator records the install; auto-uninstall runs
   asynchronously when the active count reaches zero.
3. **Queue-driven downloads (amends ADR-0004).** All downloads — initial backfill and
   new chapters — flow through an overlay `download_job` queue paced by a single
   worker, so throttling is global. Suwayomi's updater still detects new chapters, but
   its auto-download is disabled. New-chapter jobs outrank backfill.

`series.status` adds a derived `HIATUS` (no chapter for > 1 month, not completed),
recomputed weekly; the "reprise" badge is event-driven via `series.resumed_at`
(set on a new chapter after a > 1-month gap, shown for 7 days). Extension health is
tracked on `extension` plus an `extension_error_log` history for admins.

## Consequences

- The feed and read/unread colouring become SQL joins instead of N Suwayomi calls.
- We own catalogue-sync, download-pacing, and extension-health code that Suwayomi
  partly provided — the trade for global control and feed performance.
- ADR-0004's "Suwayomi's downloader runs" is superseded: Suwayomi downloads only when
  our queue tells it to.
- Spec of record: `docs/superpowers/specs/2026-06-20-overlay-data-model-design.md`.
