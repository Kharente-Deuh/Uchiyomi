// SPDX-License-Identifier: AGPL-3.0-or-later

import type { StringSchema } from 'yup'
import { string } from 'yup'
import { ACCOUNT_NAME_MAX, ACCOUNT_NAME_MIN, ACCOUNT_NAME_PATTERN } from '#shared/dto/identity/account-name'

// Mirrors the server zod rule for client-side UX validation. The server remains
// authoritative (it normalizes again before lookup/persist).
export function accountNameRule(label: string): StringSchema<string> {
  return string()
    .trim()
    .lowercase()
    .min(ACCOUNT_NAME_MIN)
    .max(ACCOUNT_NAME_MAX)
    .matches(ACCOUNT_NAME_PATTERN)
    .required()
    .label(label)
}
