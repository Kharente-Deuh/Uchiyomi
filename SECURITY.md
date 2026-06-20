# Security Policy

Uchiyomi handles authentication, sessions, and password hashing, so security
reports are taken seriously.

## Reporting a vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, use GitHub's private vulnerability reporting
("Security" tab → "Report a vulnerability") so the issue can be triaged and
fixed before disclosure. Include affected version, reproduction steps, and
impact assessment if possible.

We aim to acknowledge reports within a few days and to coordinate a fix and
disclosure timeline with you.

## Scope

The application (Nuxt/Nitro app) is the only network-exposed component.
Suwayomi-Server and PostgreSQL are internal-only by design and must never be
exposed directly.
