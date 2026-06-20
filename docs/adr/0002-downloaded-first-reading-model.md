# 2. Downloaded-first reading model

- Status: Accepted
- Date: 2026-06-20

## Context

Two reading models were possible: (a) online-streaming-first, reading directly
from sources at read time; or (b) auto-download then read the downloaded content.
The primary use case is subscribing to series and auto-downloading new chapters
to a NAS, then reading them — often offline-capable.

## Decision

Adopt a **downloaded-first** reading model. Suwayomi downloads chapters (as CBZ +
`ComicInfo.xml`); pages are served by Suwayomi's API and **proxied by the App's
BFF**. Per-user reading progress is recorded on each page turn.

## Alternatives considered

- **Online-streaming-first** — rejected for the primary use case: it does not match
  the "auto-download to NAS" goal and adds runtime dependency on source
  availability for every read.

## Consequences

- Everything read is stored on disk → storage must be sized on the NAS.
- The page-serving proxy is the one latency-sensitive path → it must be cached.
- Online browsing/adding still happens (via Suwayomi), but reading is from the
  downloaded library.
- CBZ + `ComicInfo.xml` output keeps files compatible with OPDS and other readers,
  not just Uchiyomi.
