# 13. Application structure: atomic-design frontend, DDD server

- Status: Accepted
- Date: 2026-06-20
- Revised: 2026-06-21 (server layer + frontend data-access layer — see "Revision" below)
- Revised: 2026-06-25 (composition root moved into domain; use-case barrel; shared import rule — see "Revision" below)

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

#### Frontend data access — services / composables / stores

The frontend talks to the backend through three cooperating layers; **no component
calls `$fetch` ad hoc**. The dependency points inward, mirroring the server:
components → composables → services → HTTP/DTOs.

- **`services/<domain>.ts`** — the only place that knows route paths. Typed HTTP calls
  against the **shared DTOs** (`#shared/dto/...`): request bodies typed against the
  request DTO, responses as the response DTO. Stateless functions over `$fetch`, no
  reactivity. e.g. `authService.login(body: LoginRequestDto): Promise<UserDto>`.
- **Composables** — reactive orchestration over a service (`loading`/`error`,
  `useUserSession`, optimistic updates); this is what components consume. Global ones
  live in `composables/use<Domain>.ts`, feature-scoped ones in
  `features/<feature>/composables/`.
- **`store/<domain>.ts`** — Pinia, reserved for genuinely **global** state (current
  user, capabilities, theme). Not every domain needs a store: prefer a composable
  until shared cross-route state is real.

Because the DTOs are the exact types the server produces via its presenters (the
`shared/` layer), the wire contract cannot drift between the two sides.

This convention is **exercised and refined in roadmap M3.1** (front data-access layer),
first on the identity/session slice — the way the server idioms below were settled by
the auth spine (M2.2).

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
  UsersRepository`. The category names the concern (`persistence`, `transport`,
  `security`, …); the tech folder names the implementation (`prisma`, `graphql`,
  `scrypt`, `memory`). Mappers (`*-repository.mapper.ts`, `toDomain`) build
  domain Models from raw rows/responses.
- **Composition root** (`application/index.ts` inside the domain): reads
  `runtimeConfig`, instantiates the adapters + use cases, and exports the
  ready-to-call singletons the routes consume. May also export shared
  infrastructure singletons and small domain helpers. The old
  `server/utils/<domain>.ts` wiring files were removed; `server/utils/` now
  holds **pure helpers only** (e.g. `account-name.ts`, `prisma.ts`).
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
├── api/                                          # Nitro routes (interface layer) — thin
├── shared/use-case.ts                            # IUseCase<Opts, Result>
├── domains/[domain]/[sub-domain]/
│   ├── <entity>.domain.ts                        # rich class Model + ports + *Params/*Opts/*Result
│   ├── application/
│   │   ├── usecases/<name>.use-case.ts           # class XUseCase implements IUseCase<XUseCaseOpts, XUseCaseResult>
│   │   ├── usecases/index.ts                     # barrel: export * from './x.use-case'
│   │   └── index.ts                              # composition root: wires + exports pre-built singletons
│   └── infrastructure/<category>/<tech>/         # class adapters implementing the ports + mappers
├── middleware/ plugins/                          # Nitro cross-cutting (e.g. session middleware)
└── utils/                                        # pure helpers only (account-name.ts, prisma.ts, …)

shared/dto/[domain]/*.dto.ts                      # wire contract (DTOs) shared by server + client
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

## Revision — 2026-06-21 (frontend data-access layer)

The original Frontend section named the `services/`, `composables/`, and `store/`
directories but left their responsibilities and relationship undefined. The roadmap
rework that integrated the frontend milestones (a dedicated "Frontend foundations"
phase) made the data-access layer an explicit foundation milestone (M3.1), so the
Frontend section above now spells out the convention: **services** (typed HTTP over
the shared DTOs, the only place that knows route paths) → **composables** (reactive
orchestration) → **stores** (Pinia, global state only), with components depending
inward and the shared DTOs guaranteeing the wire contract can't drift.

As with the server idioms, the specifics are **exercised and refined when M3.1 lands**
(first on the identity/session slice); this revision records the agreed shape up front
because it gates every feature UI. No tooling change.

### Service error handling — Result type, never throw (settled in M3.1)

Services do **not** throw. Every service method returns a discriminated
**Result**, so the wire/error boundary is explicit and callers cannot reach the
data without first handling failure:

```ts
// app/utils/api/api-error.ts
export type ApiResponse<T> =
  | { success: true, data: T }
  | { success: false, error: ApiError }
```

- The service `try`/`catch`es `$fetch` and maps any failure to a normalised
  `ApiError` (via `ApiError.fromFetchError`), returning `{ success: false, error }`;
  success returns `{ success: true, data }`. `$fetch` is called **without an
  explicit generic** — the response type is inferred from Nitro's typed routes and
  constrained by the method's `ApiResponse<T>` return type (no duplicated contract).
- The API layer lives in `features/<Feature>/api/<feature>.api.ts` and exports a
  **stateless factory** named `create<Feature>Api()` (not `use…`, which is reserved
  for reactive composables) returning an object typed by a `<Feature>Api` interface.
- **Composables** consume the Result: they branch on `res.success`, push
  `res.error` into a reactive `error` ref, and write `res.data` to the store
  through its actions. The reactive layer (`pending`/`error`) lives in the
  composable, not the service.

This pattern is used by **every** frontend service (no mix of throw/Result).

### Pinia store convention (settled in M3.1)

Stores are written as **setup stores** (composition API) and return their state,
getters, and actions **raw** — never wrapped in `storeToRefs`. `defineStore`'s
returned `useXxxStore` composable is the public surface; do not wrap it in a
custom function that re-exposes `storeToRefs(store)`, which would re-wrap getters
into refs (breaking auto-unwrap and template rendering) and drop the real Pinia
store API (`$patch`, `$reset`, actions).

```ts
// app/features/auth/store/auth.store.ts
export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserDto>()
  const isAuthenticated = computed(() => user.value !== undefined)
  const capabilities = computed(() => ({ /* … */ }))

  function setUser(value: UserDto): void { user.value = value }
  function clear(): void { user.value = undefined }

  return { user, isAuthenticated, capabilities, setUser, clear }
})
```

- **Mutation goes through named actions** (`setUser`, `clear`) — the single
  intention-revealing write point. Consumers (composables/components) never poke
  state refs from the outside (`store.user.value = …`); that scatters mutation and
  couples callers to the store's internal ref shape.
- **Reactivity at the consumer:** reading via the instance (`store.isAuthenticated`)
  is reactive (the store is a reactive proxy); **destructuring the store directly
  breaks reactivity**. To destructure state/getters while keeping it, use
  `storeToRefs` *at the call site* — `const { user, isAuthenticated } = storeToRefs(store)`.
  **Actions** destructure directly (`const { setUser, clear } = store`) — they are
  already bound and need no `storeToRefs`.
- **Stores hold state only; they do not fetch.** Fetching/orchestration lives in
  the composable, which writes to the store through its actions (the dependency
  points inward: composable → store).

## Revision — 2026-06-21 (feature-colocated data access)

The original Frontend section placed the data-access layer in three **top-level**
directories (`services/`, `store/`, `composables/`). Building M3.1 showed this
scatters one domain's slice across the tree. We instead **colocate the trio inside
the feature** (matching the team's backoffice front), so everything for a domain
lives together. Top-level `services/` and `store/` are **removed**; only genuinely
global composables (`useTheme`, `useModal`) stay in a top-level `composables/`.

```
app/
├── components/ Atom|Molecule|Organism      # pure, reusable UI (unchanged)
├── composables/                            # ONLY app-global composables
├── utils/                                  # pure utils (e.g. utils/api/api-error.ts: ApiError + ApiResponse)
└── features/<Feature>/                     # PascalCase, e.g. Auth/
    ├── api/<feature>.api.ts                # create<Feature>Api() — typed $fetch + Result
    ├── store/<feature>.store.ts            # use<Feature>Store — Pinia setup store
    ├── composables/<feature>.composable.ts # use<Feature>() — reactive orchestration
    ├── components/ modals/ constants/      # feature-scoped UI (as needed)
    └── *.test.ts                           # co-located, next to each file
```

**Filename role-suffixes:** `*.api.ts`, `*.store.ts`, `*.composable.ts` (mirrors the
server's `*.use-case.ts` / `*.domain.ts` style and the backoffice front).

**Auto-import (wired in `nuxt.config.ts`):**
- Composables — `imports.dirs: ['features/**/composables']` (Nuxt scans these dirs;
  the auto-imported name is the exported function, e.g. `useAuth`).
- Pinia stores — `pinia.storesDirs: ['features/**/store']` (paths are resolved
  relative to the app dir by `@pinia/nuxt`, **not** prefixed with `app/`).
- The **API layer is not auto-imported** — composables import their `create<Feature>Api`
  explicitly; `ApiError`/`ApiResponse` come from `~/utils/api`.

The dependency direction (`components → composables → api`, with the store as
inward global state) and the Result/store conventions above are unchanged; only the
*location* moves from top-level dirs into the feature.

## Revision — 2026-06-25 (composition root + use-case wiring pattern)

Building the extensions domain (M4.1a) settled the last open idiom: where the
composition root lives and how the barrel is structured.

**What changed:**

- **Composition root moves into the domain.** The old `server/utils/<domain>.ts`
  wiring files (`utils/extensions.ts`, `utils/identity.ts`,
  `utils/suwayomi-settings.ts`) were removed and replaced by
  `application/index.ts` inside each domain slice. This keeps all wiring
  co-located with the domain it serves and removes the cross-layer coupling from
  `server/utils/` into `server/domains/`. `server/utils/` now holds **pure
  helpers only** (e.g. `account-name.ts`, `prisma.ts`).

- **Use-case barrel.** A sibling `application/usecases/index.ts` re-exports every
  use-case module (`export * from './x.use-case'`). The composition root imports
  from this barrel; routes never import individual use-case modules directly.

- **Naming convention locked down.** Class `XUseCase`, opts type `XUseCaseOpts`
  (or `…Params`), result type `XUseCaseResult`. Constructor dependencies are
  `private readonly`; policy primitives (TTL, rate-limit window, a
  `now: () => Date` clock for testability) are injected there rather than via
  `runtimeConfig` in the use case itself.

- **Routes consume singletons.** `server/api/**` handlers import the pre-wired
  singleton directly from `~~/server/domains/<domain>/application` and call
  `.execute(opts)`. No wiring inside the route handler.

- **Presenters are pure functions** named `toXDto(...)` located under
  `infrastructure/transport/http/`. They receive a domain Model and return the
  shared DTO; they never import Prisma or Suwayomi.

- **`shared/` relative-import rule.** Files tested by the Vitest `node` project
  (which has no path aliases) **must** import from `server/shared/` using a
  **relative path** — not `~~` or `#shared`. The canonical example is the comment
  in `server/utils/account-name.ts`. This applies to any utility or use-case
  that imports `IUseCase` or any other shared type while also being covered by
  `node`-project tests.

The Server section and directory tree above have been updated to reflect this.
