// SPDX-License-Identifier: AGPL-3.0-or-later

import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'
import type { SetupState } from '~/composables/setup.api'
import { DEFAULT_PAGE } from '~/constants'
import { SETUP_PATH, statusRedirect } from '../routes'

export function requiresSetupCheck(to: RouteLocationNormalized): boolean {
  return to.name !== 'status'
}

export function resolveSetupGuard(to: RouteLocationNormalized, state: SetupState): RouteLocationRaw | undefined {
  if (state === 'unknown') {
    return statusRedirect(to)
  }

  if (to.name === 'setup') {
    return state === 'done' ? DEFAULT_PAGE : undefined
  }

  return state === 'required' ? SETUP_PATH : undefined
}
