# 6. Local accounts + SSO with revocable sessions and pragmatic RBAC

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi needs real user management: classic local accounts (username/email +
password), not HTTP Basic auth, optionally coexisting with SSO, plus an admin
console and role/permission management. Suwayomi provides only server-level auth,
not per-user accounts, so authentication is entirely Uchiyomi's concern. The
app↔Suwayomi link is service-to-service; end users never receive a Suwayomi token.

## Decision

- Use **`nuxt-auth-utils`** (scrypt password hashing, sealed cookie sessions,
  OAuth/OIDC helpers) + **`nuxt-authorization`** for abilities.
- Separate **identity from person**: `app_user` (the person) and `auth_identity`
  (provider `local` | `oidc`, subject, optional `password_hash`). A user may have
  several identities → local and SSO coexist.
- **Revocable sessions**: the sealed cookie carries only a `session.id`; the
  `session` table is the source of truth. Disabling a user kills their sessions
  immediately. The middleware loads the user (role/capabilities) per request.
- **Pragmatic RBAC**: a `role` (admin/user) plus a few capability flags
  (`can_manage_extensions`, `can_download`, `allow_nsfw`) rather than a full
  role/permission table. Controlled registration (invite/admin-created), no open
  sign-up.

## Alternatives considered

- **HTTP Basic auth / rely on Suwayomi auth** — rejected: not real multi-user, no
  per-user state, poor UX.
- **Stateless JWT only** — rejected: hard to revoke; immediate disable is required.
- **Full RBAC tables up front** — rejected: over-engineering for a household;
  revisit if custom roles are ever needed.

## Consequences

- One DB read per authenticated request (cheap vs the image proxy) enables instant
  revocation and per-request authorization.
- Content restriction (`allow_nsfw`) is kept from the start because it is hard to
  retrofit (Suwayomi exposes no clean rating; some tagging will be needed).
- Security must-dos apply: login rate-limiting, `httpOnly`/`Secure`/`SameSite`
  cookies, `NUXT_SESSION_PASSWORD` out of the repo, no secrets in logs.
