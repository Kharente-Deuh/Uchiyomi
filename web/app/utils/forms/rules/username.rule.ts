// SPDX-License-Identifier: AGPL-3.0-or-later

import type { StringSchema } from 'yup'
import { string } from 'yup'

const ACCOUNT_NAME_MIN = 3
const ACCOUNT_NAME_MAX = 32
const ACCOUNT_NAME_PATTERN = /^[a-z0-9_-]+$/

export function usernameRule(label: string): StringSchema<string> {
  return string()
    .trim()
    .lowercase()
    .min(ACCOUNT_NAME_MIN)
    .max(ACCOUNT_NAME_MAX)
    .matches(ACCOUNT_NAME_PATTERN)
    .required()
    .label(label)
}
