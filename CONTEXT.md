# Uchiyomi

Self-hosted manga and webtoon server. Users browse external sources, add titles to their personal library, and read downloaded chapters.

## Language

**Comic**:
A title fetched from an external source (manga, manhwa, manhua, etc.). Stored once in the database regardless of how many users add it.
_Avoid_: Manga (when referring to the database entity — too narrow), Series

**Library entry**:
The link between a user and a comic they have added. Each user has their own library entries; reading history will attach here.
_Avoid_: Library comic, User comic

**Add to library**:
The user-facing action that creates a comic in the database (if it does not exist yet) and creates a library entry linking the user to it. Idempotent: adding a comic already in the user's library succeeds without creating a duplicate entry. When the comic is new, obtains Cover (local) and fetches the full chapter list from the source — if either fails, the entire action fails and no comic or library entry is created. When the comic already exists (another user added it first), only the library entry is created; the existing local cover and chapter records are reused and undownloaded chapters are enqueued if not already in the queue.

**Source**:
An external site from which comics can be browsed and added. AsuraScans is the first source.

**Challenge solver**:
An optional service that solves a bot-protection challenge on a source (typically a Cloudflare interstitial) so the server can obtain the real page. Distinct from Source. Not required for sources that expose a public API.
_Avoid_: FlareSolverr (when meaning Uchiyomi's sidecar), Byparr (when meaning the role), flare solver, scraper

**Source slug**:
The identifier a source uses for a comic in its URLs (e.g. `solo-leveling` on AsuraScans). Together with the source name, it uniquely identifies a comic in the database.

**Comic status**:
The publication state of a comic as reported by its source (e.g. ongoing, completed, hiatus). Stored as-is from the source; no cross-source normalization yet.

**Comic type**:
The format of a comic as reported by its source (e.g. manga, manhwa, manhua, mangatoon). Stored as-is from the source.

**Cover (cached)**:
A temporarily stored cover image used while browsing a source, before a comic exists in the database. May expire and be discarded. When that comic is first created, a cached cover for the same source and slug is moved (not copied) to become Cover (local).

**Cover (local)**:
The permanent cover image of a comic, written only when the comic is first created. Stored as `cover.{ext}` in that comic's download directory (`downloads/{comicId}/`), next to its chapter files. A comic row exists only if this file exists: if the cover cannot be moved from cache or downloaded, the comic is not created.
_Avoid_: cache/covers (that directory is Cover (cached) only)

**Remove from library**:
Deletes only the user's library entry and that entry's reading progress. The comic and its local cover remain as long as at least one other user still has it in their library. When the last user removes a comic and it is deleted from the database, all associated chapter records, downloaded chapter files, and the local cover are deleted as well. Re-adding the comic creates a new library entry with no reading progress.

**Library**:
The personal collection of comics a user has added. Displayed on the `/library` page. Filterable by source, type, status, and a case-insensitive search on title and alt titles. Sortable by title or date added, ascending or descending. Default sort is title ascending. Each library comic is shown with its local cover.

**Feed**:
The user's library comics that have at least one unlocked chapter, listed with their latest unlocked chapters, newest first, ranked by chapter availability date. Comics with only locked chapters are omitted until the first unlock. Displayed on `/feed`. `/` redirects to `/feed`. This is the page the user lands on after login. Filterable by source, type, and status. Each shown chapter includes its download progress and whether the user has reading progress for that chapter. Tapping the card (cover or title) opens the library comic page. Tapping the shown chapter opens the chapter reader if that chapter is fully downloaded; otherwise the user stays on the Feed.
_Avoid_: Updates, timeline, homepage

**Library comic page**:
The source-agnostic detail page for a comic already in the user's library, at `/comic/:id` (internal UUID). Same role as the AsuraScans series detail page (metadata, chapter list, live download progress), but it reads stored comics and chapters rather than the source. A Continue action opens that comic's Continue. A chapter row opens that chapter at its last page (or page 1), only if it is fully downloaded. The chapter list can enter a selection mode to mark chapters as read. A Refresh action runs a chapter list refresh for that comic when it is pollable; otherwise the action is shown but not available.
_Avoid_: Series page (when referring to the library-owned page)

**Comic metadata**:
A snapshot of information returned by the source. Most fields stay as stored at add-to-library. Status and chapter count are refreshed by chapter list refresh for pollable comics.

**Reading history**:
A per-user event log of chapters read (chapter + read date), attached to a library entry. Used for a future history page. Distinct from reading progress. Not implemented in this milestone, but the data model must account for it.

**Reading progress**:
The furthest page a user reached in a given chapter of a library entry. Resume uses that page when the same chapter is opened again. A save with a lower page leaves the stored page unchanged, but that chapter still becomes Continue. Distinct from reading history. There is no read/unread flag: the list treats any stored progress as opened, and mark chapters as read is a shortcut that sets the page to the chapter's last page. Deleted with the library entry.

**Mark chapters as read**:
A library comic page action that sets reading progress to the last page (`pagesNb`) of each selected chapter whose page count is known. Not a read/unread flag and not a reading history event. Chapters with no page count yet are skipped. Locked and not fully downloaded chapters can be marked.
_Avoid_: mark as unread, read flag, bulk retry

**Continue**:
The chapter of a library entry whose reading progress was updated most recently, and the page to resume at. Derived from reading progress; not stored separately. Absent when the user has no reading progress for that comic. Mark chapters as read never moves Continue to an earlier chapter: it creates Continue when none exists, or moves it only to a later selected chapter that was marked read. A single-chapter progress save is unchanged.
_Avoid_: resume pointer, last read

**Chapter reader**:
The in-app view that displays the local page images of a fully downloaded chapter, at `/comic/:id/chapter/:number`. The URL updates when the reader moves to another chapter. It uses that user's reader settings for the comic's stored comic type. Incomplete or failed downloads cannot be opened. Changing chapter happens from the overlay (previous, next, or chapter menu), not by scrolling into another chapter.
_Avoid_: Viewer, player

**Page rail**:
A vertical overlay control on the chapter reader that shows position in the current chapter and lets the user jump or scrub to a page, or to a double-page pair when double page is on. Visible only with the overlay, including on the between-chapters card (then pinned at the first or last page). Top is the first page, bottom is the last, in every reading mode. The label uses real 1-based page numbers (a double-page pair shows both). Distinct from reading progress and from chapter download progress.
_Avoid_: Progress bar, page indicator, seek bar, slider

**Next chapter**:
The adjacent later instalment of the same comic, ordered by chapter number. Includes locked and not fully downloaded chapters; the chapter reader does not skip over them. Distinct from Continue.
_Avoid_: next readable, next fully downloaded (when that would skip a gap)

**Previous chapter**:
The adjacent earlier instalment of the same comic, ordered by chapter number. Same inclusion rules as Next chapter.

**Reading mode**:
The layout used by the chapter reader for a given user and comic type (paged right-to-left, paged left-to-right, or webtoon). One mode per type per user; not chosen per comic and not from inside the chapter reader. Webtoon is a vertical strip of the current chapter only: reaching the last page shows the overlay so the user can change chapter; the next chapter is not appended.
_Avoid_: infinite scroll (in the chapter reader)

**Reader settings**:
A user's reading preferences for one comic type: reading mode, page scale, and for paged modes whether to show a single page or a double page. Shared by every comic of that type for that user; not per comic and not per device. Factory defaults: manga is paged right-to-left, manhua paged left-to-right, both fit to screen and single page; manhwa and mangatoon are webtoon, fit to width.

**Reader settings page**:
The page at `/settings/reader` where a user selects one comic type, edits that type's reader settings, and saves. Switching the selector shows the other type's settings. Switching type with unsaved edits asks to save, discard, or cancel.
_Avoid_: four settings panels (the page shows one type at a time)

**Page scale**:
How a page image is sized in the chapter reader: fit to width, fit to height, or fit to screen.

**Chapter title**:
The optional name the source gives a chapter, stored on the chapter. Distinct from the chapter number. When empty, the UI shows the number only.

**Browse source**:
Search and explore comics on an external source before adding them. The AsuraScans browse page supports title search, sort (popular, latest update), status and type filters. Genre filters are out of scope. Clicking a card on a source browse page opens that source's series page. Cards on Library and Feed open the library comic page. Unifying browse and library comic pages later is out of scope.

**Sources page**:
Lists all available sources as cards. Each card links to that source's browse page. Only AsuraScans for now.

**Search pagination**:
Mobile: infinite scroll — next page loads when the user reaches the bottom, previous results stay visible. Desktop: classic pagination footer with page numbers. Applies to source browse, library, and feed.

**Chapter**:
A single instalment of a comic, stored in the database and linked to one comic. Chapters are created when a comic is added to a library, and when a chapter list refresh finds source chapters with no matching source chapter slug. Each chapter tracks its download progress independently. The expected page count (`pagesNb`) is initialized from the source chapter list and updated when page URLs are fetched if the source reports a different count.
_Avoid_: Episode (unless the source uses that term)

**Chapter list refresh**:
A refetch of a pollable comic's infos and chapter list from its source: creates missing chapter records, enqueues downloadable ones, and updates comic status and chapter count. Pollable comics are those whose stored status is ongoing or hiatus. Runs periodically for all pollable comics, and on demand for one pollable comic by a user who has it in their library.
_Avoid_: catalog sync, crawl, scrape, metadata refresh

**Source chapter slug**:
The identifier a source uses for a chapter in its API (e.g. the `slug` field returned by AsuraScans). Required to fetch page images from the source. Distinct from the chapter number shown to readers.
_Avoid_: Chapter URL (the public reader URL uses the chapter number, not this slug)

**Chapter download progress**:
A percentage (0–100) of page images successfully downloaded for a chapter. At 100 the chapter is fully available for offline reading. Set to -1 when a download error occurs — after 3 failed attempts per page. On server startup, chapters with progress between 0 and 99, or with a previous error (-1), are queued again automatically. Interrupted downloads resume incrementally — only missing pages are fetched.

**Chapter published at**:
The date and time the source reports the chapter was published. Stored as returned by the source.

**Chapter early access until**:
The date and time until which a chapter is locked behind early access on the source. Chapters with a future early access date are stored in the database but not queued for download until the lock expires. A periodic scan (every 15 minutes) checks for newly unlocked chapters and queues them for download.

**Unlocked chapter**:
A chapter that is not behind early access: it has no early access date, or that date is already in the past. Only unlocked chapters appear on the Feed.

**Chapter availability date**:
The date used to rank an unlocked chapter on the Feed. If early access until is set and already past, that date is the availability date; otherwise it is the chapter's published-at date.

**Chapter files (local)**:
The page images of a downloaded chapter, stored on disk under `downloads/{comicId}/{chapterNumber}/`. Each file is named with a zero-padded index (e.g. `001.webp`, `002.png`). PNG and JPEG pages from the source are converted to lossless WebP when the resulting WebP file is smaller; otherwise the original format and bytes are retained. Each comic is identified by its internal UUID, not its title or source slug. The same `{comicId}` directory holds Cover (local) as `cover.{ext}` at its root, not inside a chapter folder.

**Retry chapter download**:
The user-facing action to re-attempt downloading a chapter. If the chapter is in error (-1), all partial files are deleted and the download restarts from 0. If the chapter was merely interrupted (0 < progress < 100), only missing pages are fetched.

**Live chapter download progress**:
Chapter download progress shown on the series chapter list while the user is viewing it, kept current until every chapter is complete or failed.
_Avoid_: WebSocket, progress stream, SSE

**Chapter download queue**:
Each source has a single global download queue processed one chapter at a time (FIFO). Chapters are enqueued in ascending chapter number order. Across multiple comics, chapters are enqueued in the order their comic was added to a library. Enqueue is idempotent: a chapter is added only if its download progress is below 100 and it is not already in the queue. Chapters at 100 % or already queued are skipped. Within a chapter, pages are downloaded in parallel, throttled by a configurable rate limit on the worker.

### Auth

**OIDC provider**:
An identity provider registered on the instance so users can sign in with it. Distinct from Source.
_Avoid_: OIDC client (the registered app at the identity provider is not this entity)

**OIDC provider display name**:
The human-facing label shown on the login button ("Sign in with …"). Distinct from OIDC provider slug.
_Avoid_: OIDC_NAME, using this as a URL identifier

**OIDC provider slug**:
The unique, URL-safe identifier of an OIDC provider, used in the login start URL. Distinct from OIDC provider display name and from Source slug.
