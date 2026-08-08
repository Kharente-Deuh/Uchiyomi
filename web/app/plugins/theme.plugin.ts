// SPDX-License-Identifier: AGPL-3.0-or-later

import type { useTheme } from 'vuetify'
import { resolveInitialTheme } from '~/utils/theme'

export default defineNuxtPlugin((nuxtApp) => {
  const stored = useCookie<string | undefined>('uchiyomi_theme', {
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
    path: '/',
  })

  nuxtApp.hook('vuetify:before-create', ({ vuetifyOptions }) => {
    const themeOptions = vuetifyOptions.theme
    if (!themeOptions || typeof themeOptions !== 'object') {
      return
    }

    themeOptions.defaultTheme = resolveInitialTheme(
      stored.value,
      Object.keys(themeOptions.themes ?? {}),
      themeOptions.defaultTheme ?? 'dark',
    )
  })

  nuxtApp.hook('app:beforeMount', () => {
    const theme = (useNuxtApp().$vuetify as { theme: ReturnType<typeof useTheme> }).theme
    watch(theme.global.name, (name) => {
      stored.value = name
    })
  })
})
