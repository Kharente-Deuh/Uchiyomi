# 13. Application structure: atomic-design frontend, DDD server

- Status: Accepted
- Date: 2026-06-20
- Revised: 2026-06-21 (server layer — see "Revision" below)

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

### Server (`server/`) — hexagonal, rich-domain, organised by (sub-)domain

Organise by domain — and, where a domain is broad, by **sub-domain** — each a
ports-and-adapters (hexagonal) slice. Nitro's file-based routes (`server/api/`)
are the **interface layer** and stay thin: validate input, resolve the session,
call a use case, map the result/errors. There is no controller class — the
route-handler file *is* the controller (path = route); business logic never
lives in `server/api/`.

Every grouping is a **flat ES module consumed via `import * as`** (no TS
`namespace` keyword). Per-layer conventions:

- **Domain** (`<entity>.domain.ts`): a rich `class Model` (data **and** behaviour
  methods, populated by `Object.assign` in the constructor with `declare` fields),
  value types, the repository/service **ports** (interfaces), and per-method
  `*Params` / `*Opts` / `*Result` shapes. The domain imports nothing from
  Nuxt/Prisma. Absence is `undefined` (never `null`).
- **Application** (`application/usecases/<name>.use-case.ts`): one use case per
  module — `export class UseCase implements IUseCase<Opts, Result>` (the shared
  port in `server/shared/use-case.ts`), dependencies injected via the
  constructor. Use cases depend only on ports; **no input validation here**.
- **Infrastructure** (`infrastructure/<category>/<tech>/`): the adapters that
  implement the ports — classes such as `PrismaUserRepository implements
  User.Repository`. The category names the concern (`persistence`, `transport`,
  `security`, …); the tech folder names the implementation (`prisma`, `graphql`,
  `scrypt`, `memory`). Mappers (`*-repository.mapper.ts`, `toDomain`) build
  domain Models from raw rows/responses.
- **Composition root** (`server/utils/<domain>.ts`): reads `runtimeConfig`,
  instantiates the adapters + use cases, and exports the ready-to-call instances
  the routes consume.
- **DTOs live in the Nuxt `shared/` layer** (`shared/dto/...`) so server and
  client share one wire contract (the client cannot import from `server/`).
  Routes map domain Models to DTOs via a server-side presenter
  (`*-http.presenter.ts`) and never expose a `Model`. Request bodies are
  validated with `zod` in the route, typed against the shared request DTO
  (`… satisfies z.ZodType<…>`).
- **Policy constants** (session TTL, password length, rate-limit windows…) live
  in `runtimeConfig`, not the domain; domain methods receive them as parameters.

```
server/
├── api/                                     # Nitro routes (interface layer) — thin
├── shared/use-case.ts                       # IUseCase<Opts, Result>
├── domains/[domain]/[sub-domain]/
│   ├── <entity>.domain.ts                   # rich class Model + ports + *Params/*Opts/*Result
│   ├── application/usecases/*.use-case.ts   # class UseCase implements IUseCase
│   └── infrastructure/<category>/<tech>/    # class adapters implementing the ports + mappers
├── middleware/ plugins/                     # Nitro cross-cutting (e.g. session middleware)
└── utils/                                   # composition roots (<domain>.ts) + shared singletons (Prisma, Suwayomi)

shared/dto/[domain]/*.dto.ts                 # wire contract (DTOs) shared by server + client
```

Domains map to the overlay model and roadmap: `identity` (split into the
`users`, `sessions`, `auth`, `password` sub-domains), `catalogue` (Suwayomi
cache), and later `extensions`, `library`, `downloads`, `reading`. Prisma and
the Suwayomi client are **infrastructure**; domain and application depend on
ports, never on Prisma/HTTP directly. See [`docs/architecture.md`](../architecture.md)
for a narrated walk-through with examples.

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

## Revision — 2026-06-21 (server layer)

The original decision named the three-layer split (domain / application /
infrastructure) but left the per-layer idioms open. Building the auth spine
(M2.2) settled them, and the Server section above now reflects the result:

- Domains are **rich classes** (`Model` with behaviour) behind **ports**, with
  **class adapters** in infrastructure — full ports-and-adapters/hexagonal.
- Everything is a **flat ES module** consumed via `import * as` (no TS
  `namespace`); use cases are `IUseCase` classes under `application/usecases/`.
- Infrastructure is organised `infrastructure/<category>/<tech>/`.
- **Input validation moves to the interface layer** (`zod` in the route), not the
  use case; **policy constants move to `runtimeConfig`**.
- **HTTP DTOs live in the Nuxt `shared/` layer** and are produced by server-side
  presenters, so domain Models never cross the wire and the client shares the
  same typed contract.
- Tooling added alongside this: `zod` (request validation), `nuxt-auth-utils` +
  `nuxt-authorization` (sessions/abilities). The frontend section is unchanged.

Rationale: the original "DDD" sketch was anaemic (interfaces + functions). The
team's other services (NestJS) use a rich-domain, ports-and-adapters style; this
aligns the Nitro backend with that shared mental model while staying idiomatic
ESM (no `namespace`, `shared/` for the wire contract, `runtimeConfig` for ops).
