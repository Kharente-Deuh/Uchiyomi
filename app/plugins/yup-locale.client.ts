// SPDX-License-Identifier: AGPL-3.0-or-later
// Wires yup's global validation messages to the app's i18n (`validation.*` keys)
// and re-applies them whenever the active locale changes. Client-only: `setLocale`
// mutates yup's global state, which must not leak across SSR requests, and form
// validation runs in the browser anyway. The wiring lives in `setupYupLocale`
// (testable without a Nuxt runtime); this plugin is just the runtime shim.
import { defineNuxtPlugin, useNuxtApp } from '#imports'
import { setupYupLocale } from '~/utils/forms/yup-locale'

export default defineNuxtPlugin(() => {
  const { t, locale } = useNuxtApp().$i18n

  setupYupLocale((key, named) => (named ? t(key, named) : t(key)), locale)
})
