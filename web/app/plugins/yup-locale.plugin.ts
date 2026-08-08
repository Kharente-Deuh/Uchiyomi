// SPDX-License-Identifier: AGPL-3.0-or-later

import { defineNuxtPlugin, useNuxtApp } from '#imports'
import { setupYupLocale } from '~/utils/forms/yup-locale'

export default defineNuxtPlugin(() => {
  const { t, locale } = useNuxtApp().$i18n

  setupYupLocale((key, named) => (named ? t(key, named) : t(key)), locale)
})
