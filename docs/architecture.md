# Backend architecture

A friendly tour of how the Nitro backend (`server/`) is organised. The formal
decision lives in [ADR-0013](./adr/0013-application-structure.md); this page is
the narrated version with examples. The frontend (`app/`) has its own shape
(atomic design) — see the ADR.

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

## See also

- [ADR-0013](./adr/0013-application-structure.md) — the structural decision (and its 2026-06-21 server revision).
- [ADR-0005](./adr/0005-nuxt-nitro-monolith.md) — why a single Nuxt/Nitro monolith.
- [ADR-0006](./adr/0006-auth-local-and-sso.md) — the auth/session model.
- [ADR-0008](./adr/0008-prisma-overlay-data-access.md) — Prisma overlay.
