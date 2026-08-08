// SPDX-License-Identifier: AGPL-3.0-or-later

import { DEFAULT_PAGE } from '~/constants'

export default defineNuxtRouteMiddleware((to) => {
  if (to.matched.length === 0) {
    return navigateTo(DEFAULT_PAGE, { replace: true })
  }
})
