// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineNuxtPlugin } from '#imports'
import { aliases, iconify } from '../iconsets/iconify'

export default defineNuxtPlugin({
  name: 'uchiyomi:vuetify-icons',
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
