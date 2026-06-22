// SPDX-License-Identifier: AGPL-3.0-or-later
// Single source of truth for the displayName bounds, shared by the server zod
// schema and the front yup rule. Pure: no zod/yup import (neither bundle pulls
// the other's lib).
export const DISPLAY_NAME_MIN = 1
export const DISPLAY_NAME_MAX = 64
