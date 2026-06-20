# 14. Background jobs and scheduling run in-process in Nitro

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi needs background work: a download queue worker (ADR-0012's queue-driven
downloads), periodic crons (weekly `HIATUS` recompute, catalogue sync, extension
auto-uninstall), and the brownfield Suwayomi import. The backend is a single
Nuxt/Nitro monolith (ADR-0005) and the deployment target is a **persistent Node
process on a NAS** — not serverless/edge. The question: a separate worker service,
or run this work inside the Nitro process?

## Decision

Run background work **in-process within Nitro**. No separate backend application
(ADR-0005 holds).

- **Download worker** — a `server/plugins/` Nitro plugin starts a queue consumer at
  boot. It dequeues `DownloadJob` with `SELECT … FOR UPDATE SKIP LOCKED`, applies a
  **global pacing** delay, retries with backoff, and is **env-gated** (off in
  dev/test so local runs never hit source sites).
- **Scheduled jobs (crons)** — Nitro **scheduled tasks** (`server/tasks/` +
  `scheduledTasks` cron, behind `experimental.tasks`) for the weekly hiatus
  recompute, catalogue sync, extension auto-uninstall, etc. They run in the same
  persistent process; no external system cron is needed.
- **Concurrency safety** — the queue uses row-level locking
  (`FOR UPDATE SKIP LOCKED`), so processing stays correct even if more than one
  worker ever runs.
- **Isolation, if ever needed** — deploy the **same `.output` build** a second time
  as a worker-only process via an env flag. Same codebase, no second app.

## Alternatives considered

- **Dedicated worker service (e.g. BullMQ + Redis, separate process)** — robust and
  scalable, but adds Redis and a second deployable; overkill for a household NAS.
  Rejected for now; the in-process design can graduate to this if scale demands.
- **External cron** (system crontab / platform scheduler hitting an HTTP endpoint) —
  required only on serverless, which we do not target.

## Consequences

- Simplest operations: one process, one image, one thing to deploy.
- Worker lifecycle is tied to the web server (restarts on deploy) — acceptable for
  the use case. Downloads are **I/O-bound** (awaiting Suwayomi), so the worker does
  not block request latency.
- Workers must be **gated off in dev/test**.
- If we ever scale horizontally, the row-lock keeps the queue safe, but scheduled
  tasks would need to run on a single instance (a leader/lock) to avoid duplicate
  cron runs — a future consideration, not a today problem on a single-instance NAS.
- Detailed shapes land with the relevant milestones: download worker (M3.3),
  crons/sync (M3.4), brownfield import (M3.5).
