// SPDX-License-Identifier: AGPL-3.0-or-later

import { requiresSetupCheck, resolveSetupGuard } from '~/utils/guards'

export default defineNuxtRouteMiddleware(async (to) => {
  if (!requiresSetupCheck(to)) {
    return
  }

  const done = useState('setup-done', () => false)

  if (done.value) {
    return
  }

  const state = await createSetupApi().getSetupStatus()

  done.value = state === 'done'

  const redirect = resolveSetupGuard(to, state)
  if (redirect) {
    return navigateTo(redirect)
  }
})
