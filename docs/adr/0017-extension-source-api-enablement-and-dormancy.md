# 17. Extension/source API surface, source enablement, and dormancy

- Status: Accepted
- Date: 2026-06-23

## Context

ADR-0012 (revised M4.1a) established admin-managed extensions, a global installed
set, and a per-user NSFW hybrid gate. The first M4.1a implementation exposed this
through an API surface that proved ambiguous and over-split:

- Two list endpoints (`GET /extensions` for users, `GET /admin/extensions` for the
  full catalogue) with diverging DTO shapes.
- A separate `POST /admin/extensions/:pkgName` for install/uninstall.
- Source operations split between `/extensions/:pkgName/sources` (list) and
  `/admin/sources/:id/preferences` (configure), making the source/extension
  boundary hard to read.

Two modelling gaps also surfaced. **Extension vs source**: an *extension* is the
installable package (`pkgName`, the install/uninstall and health unit); a *source*
is a queryable catalogue exposed *inside* an extension (`sourceId`, the
browse/search and preferences unit). One extension exposes N sources (typically one
per language). There was no way to keep an extension installed while hiding a subset
of its sources — enablement was all-or-nothing at the extension level. **Lifecycle
semantics**: nothing defined what happens to a user's followed series, reading
history, and downloaded chapters when a source is hidden or an extension is removed.

## Decision

**1. Resource-oriented API, no `/admin` prefix.** Access control is a `403` guard
*inside* the handler keyed on `canManageExtensions`, not a path segment. The
extension/source surface becomes:

| Method | Route | Access | Purpose |
| --- | --- | --- | --- |
| `GET` | `/extensions` | any | List. Admin sees all available; non-admin sees only installed. |
| `GET` | `/extensions/:pkgName` | any | Extension detail, including the full `ExtensionHealthDto`. |
| `POST` | `/extensions/:pkgName` | admin (403) | `{ action: 'install' \| 'uninstall' }`. |
| `GET` | `/extensions/:pkgName/sources` | any | Sources of the extension (filtered view). |
| `PATCH` | `/sources/:id` | admin (403) | Toggle a source's global `isEnabled`. |
| `GET` | `/sources/:id/preferences` | admin (403) | Read a source's preferences. |
| `PUT` | `/sources/:id/preferences` | admin (403) | Update a source's preferences. |

`sources` is a first-class resource; preferences live under `/sources/:id`, listing
is the nested view under the owning extension. (Existing `/admin/users/*` routes
predate this convention and are migrated separately, out of scope here.)

**2. One list endpoint, role-branched, with filters.** `GET /extensions` branches on
`canManageExtensions` and accepts query filters `?nsfw`, `?isHealthy`,
`?isInstalled`, `?hasUpdate`, applied server-side. `ExtensionDto` gains
`isHealthy?: boolean` (a lightweight projection of health; undefined for
non-installed extensions, which have no health record). The full health detail
(`consecutiveFailures`, error log, …) is served only by `GET /extensions/:pkgName`
via `ExtensionHealthDto`.

**3. Per-source enablement (global, admin).** A new overlay `source` table mirrors
the Suwayomi sources of installed extensions and carries an admin-managed global
`isEnabled` flag. Source visibility is now governed by three independent levers, in
order: (a) the owning extension is installed; (b) the source is `isEnabled`; (c) the
per-user NSFW hybrid gate (`allowNsfw && showNsfw`) for `isNsfw` sources. This adds
the missing fine-grained lever (e.g. keep a multi-language extension installed but
expose only `fr`) without resurrecting per-user activation, which stays dropped
(ADR-0012).

**4. Dormancy, never deletion.** Disabling a source (`isEnabled = false`) or
uninstalling an extension **never deletes** overlay user data — subscriptions,
cached series/chapters, and reading history are preserved. Affected series become
*dormant*: kept in the library, flagged with a **"Source coupée" / "Source offline"**
badge, frozen (no new-chapter detection, no new downloads). Already-downloaded
chapters stay readable (Suwayomi keeps downloaded files on disk and turns removed
sources into inert stubs; verified against Suwayomi-Server behaviour). Non-downloaded
chapters of a dormant series are not readable until the source is re-enabled or the
extension reinstalled, which wakes everything back up.

## Alternatives considered

- **Keep the two list endpoints.** Rejected: the user/admin split is a `403` and a
  result-set filter, not two resources; one role-branched endpoint with a unified
  DTO is simpler and removes shape drift.
- **`isEnabled` at the extension level only.** Rejected: cannot hide a single
  language of a multi-source extension; the real lever is per-source.
- **Cascade-delete library/history on disable or uninstall.** Rejected: that is
  per-user data (ADR-0004); an admin action must not destroy it. Dormancy is
  reversible and lossless.
- **Hide dormant series entirely.** Rejected: silently dropping a followed series is
  worse UX than showing it with a clear "Source coupée" badge and read access to
  what is already downloaded.

## Consequences

- Amends ADR-0012: adds the `source` overlay table and the global `isEnabled` lever;
  formalises dormancy semantics over ADR-0002's downloaded-first reading model.
- The `source` table must stay synced with Suwayomi's source set for installed
  extensions (new rows on install, marked stale/removed on uninstall) — additional
  sync surface we own.
- Source listing and the home feed must filter on the three-lever visibility rule;
  the dormant state is derived (extension uninstalled OR source `isEnabled = false`).
- `/admin/extensions` and `/admin/sources` routes are removed; `/admin/users/*`
  remains pending a separate migration to the resource-oriented convention.
