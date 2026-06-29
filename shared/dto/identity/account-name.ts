// SPDX-License-Identifier: AGPL-3.0-or-later

// Single source of truth for the account-name identifier rule (Lot A).
// Pure: NO zod / yup import — the server (zod) and front (yup) each build their
// own validator from these constants so neither bundle pulls the other's lib.
export const ACCOUNT_NAME_MIN = 3
export const ACCOUNT_NAME_MAX = 32
export const ACCOUNT_NAME_PATTERN = /^[a-z0-9_-]+$/

// Normalize before validate/persist/lookup: `Admin` and ` admin ` both become
// `admin`, so case/whitespace variants collide on the unique index (correct).
export function normalizeAccountName(raw: string): string {
  return raw.trim().toLowerCase()
}
