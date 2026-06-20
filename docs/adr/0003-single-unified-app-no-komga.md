# 3. Single unified app; do not use Komga as the user-facing layer

- Status: Accepted
- Date: 2026-06-20

## Context

A pre-built stack of Suwayomi (for sources/downloads) + Komga (for multi-user
reading) would cover many requirements with little code. However, it forces **two
interfaces**: one to manage extensions and add titles (Suwayomi), one to read with
multi-user/SSO (Komga). Komga has no concept of sources/extensions, so the
"browse and add from a source" screen cannot exist in it. The UX requirement is a
**single** modern interface for both browsing/adding and reading.

## Decision

Build a **single unified application** that wraps Suwayomi headless and adds the
user-facing experience (browse/add, library, reader, accounts). **Do not** use
Komga (or any library-only server) as the user-facing layer.

## Alternatives considered

- **Suwayomi + Komga, two UIs** — rejected: violates the single-interface UX goal;
  Komga can never host the source-browsing experience.
- **Fork Suwayomi-WebUI** (React) and add SSO/restyle — viable for a single shared
  library, but it is React (not the chosen Nuxt) and its single-user assumptions
  fight the per-user overlay; rejected in favor of a purpose-built app.

## Consequences

- More frontend work than assembling existing tools, justified by the unified UX
  and full control over design and multi-user behavior.
- Komga is dropped from the architecture entirely.
- The app must reimplement browse/library/reader screens on top of Suwayomi's API.
