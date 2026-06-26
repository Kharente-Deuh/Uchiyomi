// SPDX-License-Identifier: AGPL-3.0-or-later
import { createExtensionsApi } from '~/features/extensions/api/extensions.api'
import { useSingleExtensionStore } from '~/features/extensions/store/single-extension.store'

export default defineNuxtRouteMiddleware(async (to) => {
  const pkgName = (to.params as { pkgName: string }).pkgName
  const api = createExtensionsApi()
  const store = useSingleExtensionStore()

  const res = await api.getExtension(pkgName)
  if (!res.success) {
    return navigateTo('/browse/extensions')
  }

  if (res.success) {
    store.setExtension(res.data.extension)
    if (res.data.health) {
      store.setHealth(res.data.health)
    }
  }
})
