// SPDX-License-Identifier: AGPL-3.0-or-later
import type { StringSchema } from 'yup'
import { string } from 'yup'
import { DISPLAY_NAME_MAX, DISPLAY_NAME_MIN } from '#shared/dto/identity/display-name'

// Mirrors the server displayNameSchema for client-side validation. The server
// remains authoritative.
export function displayNameRule(label: string): StringSchema<string> {
  return string()
    .trim()
    .min(DISPLAY_NAME_MIN)
    .max(DISPLAY_NAME_MAX)
    .required()
    .label(label)
}
