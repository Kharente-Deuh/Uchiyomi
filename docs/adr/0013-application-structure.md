# 13. Application structure: atomic-design frontend, DDD server

- Status: Accepted
- Date: 2026-06-20

## Context

The codebase needs an agreed structure before feature work spreads. Two layers,
two natural shapes: the Nuxt frontend (`app/`) is presentation-heavy and benefits
from a component taxonomy that separates pure UI from domain logic; the Nitro
backend (`server/`, ADR-0005) is logic-heavy and benefits from Domain-Driven
Design so the overlay (ADR-0004/0008) and the Suwayomi client (ADR-0001) stay
behind clear domain boundaries.

## Decision

### Frontend (`app/`) — modified atomic design

Presentation components are **pure and reusable with no business logic**; domain
logic lives under `features/`.

```
app/
├── components/            # Pure, reusable — NO business logic
│   ├── Atom/              # Buttons, inputs, chips
│   ├── Molecule/          # Search bars, filter groups
│   └── Organism/          # Tables, modal wrappers
├── features/[feature]/    # Business logic per domain
│   ├── components/        # Feature-specific components
│   ├── composables/       # Feature-specific composables
│   ├── modals/            # Feature-specific modals
│   └── constants/
├── pages/                 # File-based routing (Nuxt)
├── layouts/               # e.g. DefaultLayout.vue
├── store/                 # Pinia stores, by domain
├── services/              # API layer, by domain
├── composables/           # Global composables (useModal, useTheme…)
├── types/                 # Global TypeScript types
└── utils/                 # Pure utilities (date, string, array…)
```

i18n in templates uses the global `$t`; `<script setup>` uses `useI18n().t`
(ADR-0011).

### Server (`server/`) — Domain-Driven Design

Organise by domain, each with a DDD layering. Nitro's file-based routes
(`server/api/`, `server/routes/`) are the **interface layer** and stay thin —
they validate input and delegate to application services. There is **no separate
controller construct**: Nitro has no controller classes, so the route-handler file
*is* the controller (path = route). A handler validates input, resolves the
session, calls an application use case, then maps the result/errors to a response;
business logic never lives in `server/api/`.

```
server/
├── api/                   # Nitro route handlers (interface layer) — thin
├── domains/[domain]/
│   ├── domain/            # entities, value objects, domain services, repo interfaces
│   ├── application/       # use cases / application services, DTOs
│   └── infrastructure/    # Prisma repositories, Suwayomi adapters, mappers
├── plugins/ middleware/   # Nitro cross-cutting
└── utils/                 # shared (e.g. the Prisma client singleton)
```

Likely domains (mapping to the overlay model and the roadmap): `identity`
(users/auth/sessions/invites), `extensions`, `library`
(subscriptions/categories), `catalogue` (series/chapter cache + Suwayomi),
`downloads` (queue/worker), `reading` (progress/preferences). Prisma and the
Suwayomi client are **infrastructure**; domain and application layers depend on
repository *interfaces*, not on Prisma directly.

## Alternatives considered

- **Flat `components/` + technical-layer backend** (controllers/services/models) —
  simpler initially, but blurs the pure-UI vs domain boundary on the frontend and
  scatters domain logic on the backend. Rejected for a multi-domain app.
- **Strict classical atomic design** (atoms→…→pages only) — no place for
  feature-scoped logic; `features/` is the pragmatic addition.

## Consequences

- Clear home for every unit: pure UI in `components/`, domain UI/logic in
  `features/`, server domain logic behind interfaces in `server/domains/`.
- The server uses the three-layer DDD split (domain / application / infrastructure)
  per domain; per-domain specifics emerge during implementation, starting with the
  first server milestone (the Suwayomi integration layer).
- Empty directories are not committed; folders appear as features arrive.
