# Backend architecture

A friendly tour of how the Nitro backend (`server/`) is organised. The formal
decision lives in [ADR-0013](./adr/0013-application-structure.md); this page is
the narrated version with examples. The frontend (`app/`) has its own shape —
jump to [Frontend architecture](#frontend-architecture) below.

## TL;DR

The backend is **hexagonal** (ports & adapters) with a **rich domain**:

- The **domain** describes the business — entities with behaviour (`class Model`)
  and the *ports* (interfaces) it needs from the outside world. It knows nothing
  about Prisma, HTTP, or Nuxt.
- **Use cases** orchestrate the domain to do one thing (`Login`, `CreateUser`).
  They depend only on ports.
- **Adapters** in `infrastructure/` implement the ports with real tech (Prisma,
  GraphQL, scrypt…).
- **Routes** (`server/api/`) are thin controllers: validate → call a use case →
  return a DTO.
- The wire contract (**DTOs**) lives in `shared/`, so the frontend and backend
  share the exact same types.

Everything is a plain **ES module** imported with `import * as Thing` — no TS
`namespace`.

```
            HTTP request
                │
        ┌───────▼────────┐   server/api/**            validate (zod) → map to DTO
        │   Route (thin) │   the "controller"
        └───────┬────────┘
                │ calls .execute(opts)
        ┌───────▼────────┐   application/usecases/**   orchestration, no I/O details
        │    Use case    │   depends on PORTS only
        └───────┬────────┘
                │ uses ports (interfaces)
        ┌───────▼────────┐   <entity>.domain.ts        entities + behaviour + ports
        │     Domain     │   no Nuxt / no Prisma
        └───────┬────────┘
                │ ports implemented by
        ┌───────▼────────┐   infrastructure/<cat>/<tech>/
        │    Adapters    │   Prisma repo, scrypt hasher, GraphQL repo…
        └────────────────┘
```

Dependencies point **inward**: routes → use cases → domain. Infrastructure also
points inward (it implements domain ports). The domain depends on nothing.

## The layers, with real examples

We'll use the `identity` domain (`server/domains/identity/`), which is split into
sub-domains: `users`, `sessions`, `auth`, `password`.

### 1. Domain — `*.domain.ts`

A rich entity plus the ports it needs. Example: `users/user.domain.ts`.

```ts
// consumed elsewhere as: import * as User from '.../user.domain'
export type Role = 'ADMIN' | 'USER'

export type ModelProps = Omit<Model, 'canManageUsers' | 'isActive'>

export class Model {
  declare id: string
  declare email: string
  declare role: Role
  declare passwordHash?: string        // stays here, never crosses HTTP (see DTOs)
  // …
  constructor(data: ModelProps) { Object.assign<ModelProps, ModelProps>(this, data) }

  canManageUsers(): boolean { return this.role === 'ADMIN' }   // behaviour lives on the entity
  isActive(): boolean { return this.status === 'ACTIVE' }
}

export interface FindByEmailParams { email: string }
// … one *Params / *Opts / *Result shape per port method

export interface Repository {          // the PORT — infrastructure implements this
  findByEmail: (params: FindByEmailParams) => Promise<Model | undefined>
  // …
}
```

Notes:
- **Behaviour belongs on the model** (`user.canManageUsers()`), not as free
  functions.
- Fields use `declare` so the `Object.assign` constructor type-checks under
  `strict` without disabling any compiler check. The `ModelProps` type (an
  `Omit` of the methods) is what the constructor accepts, so callers can't forget
  a field.
- Port methods take a **single `params` object** (easy to extend) and return
  `Model | undefined` — absence is always `undefined`, never `null`.

### 2. Application — `application/usecases/*.use-case.ts`

One use case per file, a class implementing the shared `IUseCase` port, with its
dependencies injected through the constructor. Example: `auth/.../login.use-case.ts`.

```ts
import type { IUseCase } from '../../../../shared/use-case'
import * as User from '../../users/user.domain'
import * as Session from '../../sessions/session.domain'
// …

export interface Opts { email: string, password: string, userAgent?: string, ip?: string }

export class UseCase implements IUseCase<Opts, Session.Model> {
  constructor(
    private readonly usersRepository: User.Repository,
    private readonly sessionsRepository: Session.Repository,
    private readonly passwordHasher: Password.Hasher,
    private readonly ttlMs: number,                       // a config value, injected
    private readonly now: () => Date = () => new Date(),  // injectable clock for tests
  ) {}

  async execute(opts: Opts): Promise<Session.Model> {
    const user = await this.usersRepository.findByEmail({ email: opts.email })
    const ok = user && user.isActive() && user.passwordHash
      && await this.passwordHasher.verify({ hash: user.passwordHash, password: opts.password })
    if (!ok || !user) {
      throw new Auth.AuthError('invalid_credentials', 'Invalid email or password')
    }
    return this.sessionsRepository.create({ userId: user.id, expiresAt: Session.newExpiry(this.now(), this.ttlMs) /* … */ })
  }
}
```

Rules: use cases **only touch ports** (never Prisma/HTTP), do **no input
validation** (that's the route's job), and receive policy values (TTL, etc.) as
constructor params.

### 3. Infrastructure — `infrastructure/<category>/<tech>/`

The adapters that make ports real, as classes implementing the port. The folder
path encodes the *concern* then the *technology*:

```
users/infrastructure/persistence/prisma/prisma-user.repository.ts          # class PrismaUserRepository implements User.Repository
users/infrastructure/persistence/prisma/prisma-user-repository.mapper.ts   # toDomain(row): User.Model
password/infrastructure/security/scrypt/scrypt-password.hasher.ts          # class ScryptPasswordHasher implements Password.Hasher
auth/infrastructure/persistence/memory/memory-login-rate-limiter.ts        # class MemoryLoginRateLimiter implements Auth.RateLimiter
catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository.ts
```

A **mapper** turns raw rows/responses into domain Models:

```ts
export function toDomain(row: UserRow): User.Model {
  return new User.Model({ id: row.id, email: row.email, /* … */ })
}
```

Categories seen so far: `persistence` (databases, in-memory state), `transport`
(GraphQL/HTTP to other services), `security` (hashing). Add new categories as new
kinds of adapter appear.

### 4. Interface — routes in `server/api/**`

The route file *is* the controller (its path is the URL). It validates the body
with `zod` (typed against the shared request DTO), calls a use case, and returns
a **DTO** — never a domain Model. Example: `server/api/auth/login.post.ts`.

```ts
import type { LoginRequestDto } from '#shared/dto/identity/auth.request'
import { login, loginRateLimiter } from '../../utils/identity'   // composition root

const Body = z.object({ email: z.string().email(), password: z.string() }) satisfies z.ZodType<LoginRequestDto>

export default defineEventHandler(async (event) => {
  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  // … rate-limit, then:
  const session = await login.execute({ email: parsed.data.email, password: parsed.data.password /* … */ })
  // … set the sealed cookie, return { ok: true }
})
```

### 5. Composition root — `server/utils/<domain>.ts`

This is where the layers are wired. It reads config, builds the adapters, and
instantiates the use cases — the routes just import the ready instances.

```ts
// server/utils/identity.ts
const { auth } = useRuntimeConfig()
const userRepository = new PrismaUserRepository(prisma)
const passwordHasher = new ScryptPasswordHasher()
export const login = new Login.UseCase(userRepository, sessionRepository, passwordHasher, auth.sessionTtlMs)
// … one export per use case
```

## DTOs & the `shared/` layer

DTOs are the **wire contract** — the only shapes that cross the HTTP boundary.
They live in `shared/dto/...` (the Nuxt 4 shared layer), so:

- the **server** maps a domain Model → DTO with a presenter
  (`*-http.presenter.ts`) and returns that;
- the **frontend** can import the exact same type (`import type { UserDto } from
  '#shared/dto/identity/user.dto'`) — and Nitro also infers it on `$fetch`/`useFetch`.

Because the frontend can't import from `server/`, and DTOs sit in `shared/`, the
domain `Model` (which may carry `passwordHash`) **never** leaves the server. The
presenter copies only safe fields:

```ts
export function toUserDto(user: Omit<User.Model, 'passwordHash'>): UserDto { /* explicit field copy, no passwordHash */ }
```

Request DTOs (`SetupRequestDto`, `LoginRequestDto`, …) also live in `shared/`; the
route's zod schema is checked against them with `satisfies z.ZodType<…>`, so the
validation and the shared type can't drift.

## Configuration

Operational/policy constants are **not** hard-coded in the domain — they live in
`runtimeConfig` (overridable per environment via `NUXT_*` env vars):

```ts
// nuxt.config.ts → runtimeConfig.auth
{ sessionTtlMs, sessionRefreshThresholdMs, minPasswordLength, loginRateLimitMaxAttempts, loginRateLimitWindowMs }
```

The composition root reads them and injects them into use cases; domain methods
that need a value take it as a parameter (e.g. `session.shouldRefresh(now, thresholdMs)`).

## Recipes

### Add a use case to an existing domain
1. `application/usecases/<name>.use-case.ts`: `export interface Opts` + `export class UseCase implements IUseCase<Opts, Result>`; inject the ports you need.
2. Wire it in the domain's composition root (`server/utils/<domain>.ts`).
3. Add/extend the route in `server/api/**` (validate with zod against a shared request DTO; return a DTO).
4. Unit-test the use case with in-memory fakes of the ports.

### Add a new domain
1. `server/domains/<domain>/<entity>.domain.ts`: the `Model`(s), value types, ports, `*Params/*Opts/*Result`.
2. `infrastructure/<category>/<tech>/…`: adapter classes implementing the ports + mappers.
3. `application/usecases/…`: the use cases.
4. `shared/dto/<domain>/…`: the DTOs; a presenter under the domain's `infrastructure/transport/http/`.
5. `server/utils/<domain>.ts`: the composition root.
6. Routes under `server/api/<domain>/…`.

## Testing

- **Domain & use cases**: pure unit tests with in-memory fakes of the ports (no DB). Inject a fixed clock where time matters.
- **Adapters**: integration tests against a real test database / service (they skip when the resource isn't configured).
- **End-to-end**: HTTP tests that boot the app (`@nuxt/test-utils/e2e`) and exercise routes against the test DB.
- DB-backed test files run serially (`fileParallelism: false` on the vitest `node` project) because they share one test database.

---

# Frontend architecture

A tour of how the Nuxt frontend (`app/`) talks to the backend. The formal decision
(and its 2026-06-21 frontend revisions) lives in
[ADR-0013](./adr/0013-application-structure.md); this is the narrated version.

## TL;DR

The data-access layer is **colocated by feature** — one domain's slice lives
together under `app/features/<Feature>/`, not scattered across top-level dirs:

- **API** (`api/<feature>.api.ts`) — typed `$fetch` over the **shared DTOs**, the
  only place that knows route paths. Returns a **Result** (`ApiResponse<T>`), it
  **never throws**. A stateless `create<Feature>Api()` factory.
- **Composable** (`composables/<feature>.composable.ts`) — reactive orchestration
  (`loading`, store hydration); the unit components consume. Auto-imported.
- **Store** (`store/<feature>.store.ts`) — Pinia setup store, **global state only**
  (current user + capabilities). Auto-imported.

Pure reusable UI stays in `components/Atom|Molecule|Organism`; truly app-global
composables (`useTheme`, `useModal`) in the top-level `composables/`. Failures are
normalised to a front-only `ApiError` (`utils/api/`).

```
        component (page / form)
             │ consumes
        ┌────▼─────────┐   features/<F>/composables/*.composable.ts
        │  Composable  │   reactive: loading, orchestration, store hydration
        └────┬─────────┘
             │ calls (Result)            ┌──────────────┐  Pinia, global state
             ▼                           │    Store     │◄─ writes via actions
        ┌──────────────┐   features/<F>/api/*.api.ts   └──────────────┘
        │     API      │   typed $fetch over #shared DTOs → ApiResponse<T>
        └────┬─────────┘
             │ $fetch
             ▼  /api/** (Nitro routes)
```

Dependencies point **inward**: component → composable → API; the store is inward
global state the composable writes through its actions. Because the API is typed on
the **same `shared/` DTOs** the server presenters produce, the wire contract can't
drift.

## The layers, with real examples

We'll use the **Auth** feature (`app/features/Auth/`), the identity/session slice.

### 1. API — `<feature>.api.ts` (Result, never throws)

The only place that knows route paths. `$fetch` is called **without an explicit
generic** — the response type is inferred from Nitro's typed routes and constrained
by the method's `ApiResponse<T>` return type.

```ts
// app/features/Auth/api/auth.api.ts — consumed as: createAuthApi()
export interface AuthApi {
  login: (body: LoginRequestDto) => Promise<ApiResponse<void>>
  me: () => Promise<ApiResponse<UserDto>>
  // … setup, logout, isSetupStatusRequired
}

async function me(): Promise<ApiResponse<UserDto>> {
  try {
    const res = await $fetch('/api/auth/me')          // typed by Nitro: { user: UserDto }

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }   // normalise, don't throw
  }
}

export function createAuthApi(): AuthApi { return { /* login, me, … */ } }
```

### 2. Composable — `<feature>.composable.ts`

Reactive orchestration over the API: owns `loading`, hydrates the store, exposes
computed views of global state. Consumes the auto-imported store and
`nuxt-auth-utils`' `useUserSession()` (used **only** as a cookie signal — the sealed
cookie carries just `{ sessionId }`, so the user comes from `/api/auth/me`).

```ts
// app/features/Auth/composables/auth.composable.ts — consumed as: useAuth()
export function useAuth(): AuthComposable {
  const authStore = useAuthStore()        // auto-imported Pinia store
  const session = useUserSession()        // cookie signal only
  const loading = ref(false)

  async function login(body: LoginRequestDto): Promise<ApiResponse<void>> {
    loading.value = true
    const res = await authApi.login(body)
    if (res.success) {
      await fetchMe()                      // login returns { ok } → hydrate user via /me
    }
    loading.value = false

    return res                             // Result flows out; the caller decides (e.g. toast on failure)
  }

  return { user: computed(() => authStore.user), isAuthenticated: computed(() => authStore.isAuthenticated), loading, login /* … */ }
}
```

The composable returns the **Result** rather than throwing or surfacing an in-page
error ref, because errors are shown as **toasts** (an imperative side-effect the
caller — or a future centralised handler — triggers from `res.error`). `loading`
stays reactive for button/spinner binding.

### 3. Store — `<feature>.store.ts` (Pinia, global state only)

A **setup store** returning state, getters, and actions **raw** (never wrapped in
`storeToRefs`). It holds state; it does **not** fetch — the composable writes to it
through named actions.

```ts
// app/features/Auth/store/auth.store.ts — consumed as: useAuthStore()
export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserDto>()
  const isAuthenticated = computed(() => user.value !== undefined)
  const capabilities = computed(() => ({ /* canManageExtensions, canDownload, allowNsfw */ }))

  function setUser(value: UserDto): void { user.value = value }
  function clear(): void { user.value = undefined }

  return { user, isAuthenticated, capabilities, setUser, clear }
})
```

**Mutation goes through actions** (`setUser`/`clear`) — never poke `store.user.value`
from outside. **Reactivity at the consumer:** read via the instance
(`store.isAuthenticated`) is reactive; to *destructure* state/getters use
`storeToRefs(store)` at the call site; actions destructure directly.

### 4. Component (the consumer)

Components stay dumb — they consume the composable, bind `loading`, and act on the
returned Result:

```vue
<script setup lang="ts">
const { login, loading } = useAuth()

async function onSubmit(): Promise<void> {
  const res = await login(form.value)
  if (res.success) {
    await navigateTo('/')
  }
  // else: show res.error via a toast (M3.2 snackbar infra)
}
</script>
```

## Errors — `ApiError` + `ApiResponse`

`utils/api/api-error.ts` holds a front-only normalised error and the Result type
(neither is a wire contract — the wire stays H3's `{ statusCode, statusMessage }`):

```ts
export class ApiError extends Error {
  readonly status: number
  static fromFetchError(err: unknown): ApiError { /* maps ofetch FetchError → status + message */ }
}
export type ApiResponse<T> = { success: true, data: T } | { success: false, error: ApiError }
```

The API layer turns every `$fetch` failure into `{ success: false, error }`; the UI
maps `error.status` to i18n copy. The data layer never holds user-facing strings.

## Auto-import wiring

`nuxt.config.ts` registers the per-feature dirs so components use `useX()` / stores
without manual imports:

```ts
imports: { dirs: ['features/**/composables'] }       // *.composable.ts → e.g. useAuth
pinia:   { storesDirs: ['features/**/store'] }        // *.store.ts → e.g. useAuthStore (resolved relative to app/)
```

The **API layer is not auto-imported** — composables import their
`create<Feature>Api` explicitly; `ApiError`/`ApiResponse` come from `~/utils/api`.
Test files (`*.test.ts`) export nothing and are excluded from the scan.

## Testing

Front tests run in the Vitest **nuxt** project (`app/**/*.{test,spec}.ts`),
co-located next to each file:

- **API** — stub the global `$fetch` (`vi.stubGlobal`); assert path/body/Result
  unwrapping and `FetchError → ApiError` normalisation.
- **Composable** — `vi.mock` the API barrel (a `vi.hoisted` `mockApi`), mock
  `useUserSession` via `mockNuxtImport`, use a real Pinia
  (`setActivePinia(createPinia())`); assert store hydration, the silent-401 path,
  and the returned Result.
- **Store** — real Pinia; assert getters and the `setUser`/`clear` actions.

## Recipes

### Add a data slice to a feature
1. `features/<F>/api/<f>.api.ts`: `create<F>Api()` returning an `<F>Api` interface; one method per route, each returning `ApiResponse<T>`.
2. `features/<F>/store/<f>.store.ts` *(only if the state is genuinely global)*: a setup store with actions.
3. `features/<F>/composables/<f>.composable.ts`: `use<F>()` orchestrating the API + store, owning `loading`.
4. Co-locate `*.test.ts` for each; the composable/store auto-import once their dirs match the config globs.

## Shell & navigation

The app shell is a **prerendered static SPA** (`ssr: false`, with `nitro.prerender.routes: ['/']`),
cached by the service worker and instantly painted on every launch — no SSR per-request DB hit.
The client hydrates the identity store from `GET /api/auth/me` before route middleware gates
access; the middleware is **UX routing only**, not a security boundary (ADR-0016).

The shell layout (`app/layouts/default.vue`) composes two pure Organism components:
- `AppBarChrome` — an app bar (top, always visible).
- `AppNavigation` — an adaptive navigation component fed by the global `useNavigation` registry;
  renders a bottom navigation bar on mobile (`max-width: 600px`) and a navigation rail on
  desktop. Destinations (`NavItem`) are registered globally (`useNavigation(): { items }`);
  currently Home and Settings, extensible for future routes.

Auth is wired by two plugins:
- `auth.client.ts` (boot hook, `defineNuxtPlugin`) — hydrates the identity store from
  `GET /api/auth/me` before the app renders, so the guard sees the correct auth state.
- `auth.global.ts` (route middleware) — calls a pure decision function `resolveAuthGuard(path, isAuthenticated)`
  to decide whether to redirect to `/login` (protects unauthenticated visits); the guard never
  throws and returns `{ redirect?: string }`.

See [ADR-0016](./adr/0016-client-rendered-spa-shell.md) for the rendering decision,
threat model, and why `ssr: false` is the right choice for a fully authenticated PWA.

## See also

- [ADR-0013](./adr/0013-application-structure.md) — the structural decision (and its 2026-06-21 server + frontend revisions).
- [ADR-0005](./adr/0005-nuxt-nitro-monolith.md) — why a single Nuxt/Nitro monolith.
- [ADR-0006](./adr/0006-auth-local-and-sso.md) — the auth/session model.
- [ADR-0008](./adr/0008-prisma-overlay-data-access.md) — Prisma overlay.
- [ADR-0016](./adr/0016-client-rendered-spa-shell.md) — client-rendered SPA shell decision.
