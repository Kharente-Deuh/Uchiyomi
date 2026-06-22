// SPDX-License-Identifier: AGPL-3.0-or-later
import { z } from 'zod'
// Relative import (NOT `#shared`) to mirror account-name.ts and stay resolvable
// in both the Nitro runtime and the node-environment vitest project.
import { DISPLAY_NAME_MAX, DISPLAY_NAME_MIN } from '../../shared/dto/identity/display-name'

// Trim then bound. Reused by the routes that accept a displayName (self-service
// PATCH /me and admin PATCH /users/[id]); the front yup rule mirrors it.
export const displayNameSchema = z.string().trim().min(DISPLAY_NAME_MIN).max(DISPLAY_NAME_MAX)
