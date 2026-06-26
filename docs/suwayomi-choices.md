# Choices vs Suwayomi

Where Uchiyomi wraps, overrides, or diverges from Suwayomi. Keep this in sync:
add a row whenever a new decision of this kind is made (and write the ADR).

| Choice | What & why | ADR |
| --- | --- | --- |
| Headless engine, never reimplemented | Suwayomi is used only through its API; we never reimplement extensions or source parsers. | [0001](./adr/0001-use-suwayomi-as-headless-engine.md) |
| Downloaded-first reading | We read downloaded content (CBZ + `ComicInfo.xml`) proxied by the BFF, not online streaming. | [0002](./adr/0002-downloaded-first-reading-model.md) |
| Shared engine + per-user overlay | One shared Suwayomi instance; per-user library/progress live in our PostgreSQL overlay. Suwayomi's global `isRead`/`inLibrary` are never used for per-user semantics. | [0004](./adr/0004-shared-engine-per-user-overlay.md) |
| Cached catalogue | Followed series/chapters are mirrored into PostgreSQL for the feed; Suwayomi stays the source of truth, we don't reparse. | [0012](./adr/0012-catalogue-cache-and-queue-driven-downloads.md) |
| Admin-managed extensions | Extensions are installed once, globally, by an admin; no per-user activation. NSFW is a per-user hybrid gate (`allowNsfw && showNsfw`). | [0012](./adr/0012-catalogue-cache-and-queue-driven-downloads.md) |
| Queue-driven downloads | Suwayomi auto-download is disabled; all downloads flow through our paced `download_job` queue. | [0012](./adr/0012-catalogue-cache-and-queue-driven-downloads.md) |
| We own Suwayomi settings | A boot-time reconciler enforces/seeds the Suwayomi settings we depend on (e.g. `autoDownloadNewChapters = false`, `extensionRepos` seed) instead of managing them per-extension or via Docker env. | [0012](./adr/0012-catalogue-cache-and-queue-driven-downloads.md) |
| Per-source enablement + dormancy | Global per-source `isEnabled` lever; disabling a source or uninstalling an extension never deletes user data — affected series go dormant. | [0017](./adr/0017-extension-source-api-enablement-and-dormancy.md) |
