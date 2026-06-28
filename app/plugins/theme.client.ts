// SPDX-License-Identifier: AGPL-3.0-or-later

import type { useTheme } from 'vuetify'
import { resolveInitialTheme } from '~/utils/theme'

// Vuetify keeps the active theme in memory only; on a client-only SPA reload
// (ssr: false) it would reset to `defaultTheme`. Persist the user's choice in a
// cookie and restore it at boot so the selected theme survives refreshes.
export default defineNuxtPlugin((nuxtApp) => {
  const stored = useCookie<string | undefined>('uchiyomi_theme', {
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
    path: '/',
  })

  // Restore before Vuetify is created: `$vuetify` does not exist yet at plugin
  // run time, so set the initial theme on the options instead (this also avoids
  // a flash of the default theme on reload).
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

  // Persist on change once Vuetify is ready (`$vuetify` is now available).
  nuxtApp.hook('app:beforeMount', () => {
    const theme = (useNuxtApp().$vuetify as { theme: ReturnType<typeof useTheme> }).theme
    watch(theme.global.name, (name) => {
      stored.value = name
    })
  })
})
