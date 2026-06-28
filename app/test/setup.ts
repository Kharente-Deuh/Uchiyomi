// SPDX-License-Identifier: AGPL-3.0-or-later

// Vitest setup for the `nuxt` project (app/**).
//
// vuetify-nuxt-module and @nuxtjs/i18n register as async Nuxt plugins whose
// dependency chains stall silently in the Vitest nuxt environment, so their
// `app.use()` / `app.provide()` calls never execute.  Components mounted with
// `mountSuspended` then throw:
//   - "[Vuetify] Could not find defaults instance"
//   - "_ctx.$t is not a function"
//
// Fix: after setupNuxt() (which runs in the entry.mjs beforeAll), install a
// plain Vuetify instance and a minimal vue-i18n instance into the Nuxt app's
// vueApp. Vue's app.use() is idempotent — repeated installs are no-ops.
//
// This file is listed AFTER entry.mjs in vitest.config.ts so its beforeAll
// runs after setupNuxt() has already initialised the Nuxt app context.
import { beforeAll } from 'vitest'
import { createI18n } from 'vue-i18n'
import { createVuetify } from 'vuetify'
import { tryUseNuxtApp } from '#imports'
import en from '../../i18n/locales/en.json'
import fr from '../../i18n/locales/fr.json'

beforeAll(() => {
  const nuxtApp = tryUseNuxtApp()
  if (!nuxtApp) {
    return
  }

  nuxtApp.vueApp.use(createVuetify())
  nuxtApp.vueApp.use(createI18n({
    legacy: false,
    locale: 'en',
    fallbackLocale: 'en',
    messages: { en, fr },
  }))
})
