# 6. Local accounts + SSO with revocable sessions and pragmatic RBAC

- Status: Accepted
- Date: 2026-06-20
- Revised: 2026-06-21 (no-email onboarding/reset); 2026-06-22 (account-name identifier) — see "Revision" sections below

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

## Revision — 2026-06-21 (no email is sent)

Uchiyomi targets a small, self-hosted circle of users, and we decided the server
**never sends email** — no SMTP dependency, no deliverability concerns, nothing to
configure on the NAS. This reshapes the onboarding and recovery flows:

- **Onboarding = single-use invite link.** The admin generates an expirable,
  single-use invite link; the user opens it, chooses a password, and the account
  activates. The admin transmits the link **out-of-band** (chat, in person, …),
  not by email. The admin-created-account path from the original decision remains.
- **Password reset = admin-initiated single-use link.** No self-service "forgot my
  password" email. The admin generates an expirable, single-use reset link for a
  user (same mechanic as the invite), transmitted out-of-band; consuming it sets a
  new password and revokes the user's existing sessions.
- **Email verification is removed.** With no email sent, there is nothing to verify;
  `email` is an identifier, not a verified channel.

Consequences:
- A reusable single-use-token mechanism (expirable, one-time burn, constant-time
  lookup, never logged) backs both invites and resets (roadmap M7.1 → M7.2/M7.3).
- No mail adapter, queue, or templates are needed; this removes an entire
  infrastructure concern. If transactional email is ever wanted, it would be a new
  ADR, not an assumed default.
- OIDC/SSO (M7.4) is unaffected — it involves no email either.

## Revision — 2026-06-22 (account name replaces email)

The local login identifier is an **immutable, lowercase account name**, not an
email. Email is **removed** — not stored, not validated, never sent.

- **Identifier:** `account_name` (was `email`), unique, format `^[a-z0-9_-]{3,32}$`.
- **Case-insensitive:** inputs are trimmed + lowercased before validation, lookup
  and persistence, so `Admin` and `admin` are the same account. `display_name`
  remains the free, cased, human-facing name.
- **Immutable:** set once at account creation; there is no edit path (a self-service
  / admin profile-edit capability is Lot B).
- **Rule source:** one shared definition (`shared/dto/identity/account-name.ts`)
  feeds the server zod schema and the front yup rule.

This supersedes the "username/email" language in Context above: read "username/email"
as "account name". The vestigial `Invite.email` column is dropped (link-based invites
carry no email).
