# 7. License: AGPL-3.0-or-later

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi is meant to be published as open source. The maintainer wants anyone to
be free to use and modify it, while preventing it from being turned into a closed,
proprietary hosted service. A literal "no paid services" restriction is *not*
open source (the OSI definition forbids field-of-use and commercial restrictions),
which would force a "source-available" label instead.

## Decision

License Uchiyomi under **AGPL-3.0-or-later**.

## Alternatives considered

- **Source-available with a non-compete / anti-SaaS clause** (FSL, BSL,
  Elastic License v2, PolyForm) — these match a strict "no paid service" intent but
  are not open source, complicate contributions, and forfeit the open-source label.
- **MIT / Apache-2.0** — genuinely open source but allow closed proprietary
  re-hosting with no obligation to share changes.
- **AGPL-3.0** — chosen: genuinely open source (OSI-approved) *and* its network
  copyleft requires anyone hosting a modified version for others to publish the
  modified sources, which neutralizes the closed-SaaS-fork concern.

## Consequences

- The project can be marketed as open source.
- Networked forks must publish their source — the desired protection.
- AGPL can deter some corporate adoption; acceptable given the project's goals.
- No licensing entanglement with Suwayomi (consumed via API / separate container,
  not embedded). Permissive dependencies combine fine into an AGPL project.
- Add `SPDX-License-Identifier: AGPL-3.0-or-later` headers to source files. (Not
  legal advice; have load-bearing license text reviewed.)
