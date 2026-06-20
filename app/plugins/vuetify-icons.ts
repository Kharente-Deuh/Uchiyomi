// SPDX-License-Identifier: AGPL-3.0-or-later
// With `icons.defaultSet: 'custom'`, vuetify-nuxt-module steps back and lets us
// configure icons through its `vuetify:configuration` runtime hook. The Iconify
// set and its bundled Font Awesome 6 collections live in ../iconsets/iconify.
// (ADR-0009: Iconify without UnoCSS.)
import { defineNuxtPlugin } from '#imports'
import { aliases, iconify } from '../iconsets/iconify'

export default defineNuxtPlugin({
  name: 'uchiyomi:vuetify-icons',
  // Run before vuetify-nuxt-module's own vuetify plugin so our hook is registered
  // before the `vuetify:configuration` event is emitted.
  order: -30,
  parallel: true,
  setup(nuxtApp) {
    nuxtApp.hook('vuetify:configuration', ({ vuetifyOptions }) => {
      vuetifyOptions.icons = {
        defaultSet: 'iconify',
        aliases,
        sets: { iconify },
      }
    })
  },
})
