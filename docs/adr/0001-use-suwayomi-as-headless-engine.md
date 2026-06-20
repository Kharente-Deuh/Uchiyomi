# 1. Use Suwayomi-Server as a headless source engine

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi's core value depends on broad source coverage, including niche scanlation
sites. That coverage comes from the Tachiyomi/Mihon extension ecosystem, which is
intrinsically tied to the JVM/Android runtime (extensions are Android artifacts
run via a compatibility shim). Reimplementing sources in another language would
mean writing and maintaining a parser per site — unmaintainable, and it would
discard the very thing that makes the ecosystem valuable.

Suwayomi-Server already runs these extensions and exposes a GraphQL/REST API, plus
a downloader and an update scheduler.

## Decision

Run **Suwayomi-Server headless as the source/download engine**, accessed only
through its API. Do not reimplement extensions or source parsers in Uchiyomi.

## Alternatives considered

- **Reimplement the extension/source system** (e.g. in Node) — rejected: enormous,
  unmaintainable, and loses ecosystem coverage.
- **Use a downloader with built-in connectors (e.g. Tranga)** — rejected: limited,
  hardcoded source list (it cannot load Tachiyomi extensions), which is exactly the
  weak point for niche/FR sources.

## Consequences

- Uchiyomi inherits the full extension catalogue and the downloader/updater "for
  free".
- Uchiyomi is coupled to Suwayomi's API and data model; Suwayomi IDs become
  references in the overlay (mitigated by storing a natural-key fallback).
- Suwayomi is a JVM process → plan RAM accordingly on the NAS.
