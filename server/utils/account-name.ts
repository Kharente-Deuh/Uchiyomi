// SPDX-License-Identifier: AGPL-3.0-or-later

import { z } from 'zod'
// Relative import (NOT `#shared`): this file is imported by a node-environment
// vitest test, and the vitest `node` project has no `#shared` alias. A relative
// path resolves in both the Nitro runtime and the node test.
import { ACCOUNT_NAME_MAX, ACCOUNT_NAME_MIN, ACCOUNT_NAME_PATTERN } from '../../shared/dto/identity/account-name'

// Normalize THEN validate: trim + lowercase run first, so the parsed output is
// the canonical lowercased value and bounds/pattern are checked against it.
export const accountNameSchema = z
  .string()
  .trim()
  .toLowerCase()
  .min(ACCOUNT_NAME_MIN)
  .max(ACCOUNT_NAME_MAX)
  .regex(ACCOUNT_NAME_PATTERN)
