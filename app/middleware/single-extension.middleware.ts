// SPDX-License-Identifier: AGPL-3.0-or-later

import type { RouteLocationRaw } from 'vue-router'
import { createExtensionsApi } from '~/features/extensions/api/extensions.api'
import { useSingleExtensionStore } from '~/features/extensions/store/single-extension.store'

export default defineNuxtRouteMiddleware(async (to) => {
  const pkgName = (to.params as { pkgName: string }).pkgName
  const isSettingsRoute = to.name === 'browse-extensions-pkgName-settings'

  const api = createExtensionsApi()
  const store = useSingleExtensionStore()

  if (store.extension?.pkgName === pkgName) {
    return
  }

  if (isSettingsRoute && store.settings) {
    return
  }

  const extensionRes = await api.getExtension(pkgName)
  if (!extensionRes.success) {
    return navigateTo('/browse/extensions')
  }

  store.setExtension(extensionRes.data.extension)

  if (!isSettingsRoute) {
    return
  }

  const backRoute: RouteLocationRaw = `/browse/extensions/${pkgName}`
  const settingsRes = await api.getSettings(pkgName)
  if (!settingsRes.success) {
    return navigateTo(`/browse/extensions/${pkgName}`)
  }

  if (settingsRes.data.common.length === 0 && settingsRes.data.sources.every(s => s.preferences.length === 0)) {
    return navigateTo(backRoute)
  }

  store.setSettings(settingsRes.data)
})
