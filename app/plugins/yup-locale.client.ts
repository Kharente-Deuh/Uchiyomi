// SPDX-License-Identifier: AGPL-3.0-or-later
// Wires yup's global validation messages to the app's i18n (`validation.*` keys)
// and re-applies them whenever the active locale changes. Client-only: `setLocale`
// mutates yup's global state, which must not leak across SSR requests, and form
// validation runs in the browser anyway.
import { watch } from 'vue'
import { setLocale } from 'yup'
import { fr } from 'yup-locales'
import { defineNuxtPlugin, useNuxtApp } from '#imports'

export default defineNuxtPlugin(() => {
  const { $i18n } = useNuxtApp()
  const { t, locale } = $i18n

  function apply(): void {
    // Base French dictionary from yup-locales covers any message we don't override.
    if (locale.value === 'fr') {
      setLocale(fr)
    }

    setLocale({
      mixed: {
        default: t('validation.mixed.default'),
        required: t('validation.mixed.required'),
        oneOf: ({ path, values }) => t('validation.mixed.oneOf', { path, values }),
        notOneOf: ({ path, values }) => t('validation.mixed.notOneOf', { path, values }),
      },
      string: {
        length: ({ path, length }) => t('validation.string.length', { path, length }),
        min: ({ path, min }) => t('validation.string.min', { path, min }),
        max: ({ path, max }) => t('validation.string.max', { path, max }),
        matches: ({ path, regex }) => t('validation.string.matches', { path, regex }),
        email: ({ path }) => t('validation.string.email', { path }),
        url: ({ path }) => t('validation.string.url', { path }),
        trim: ({ path }) => t('validation.string.trim', { path }),
        lowercase: ({ path }) => t('validation.string.lowercase', { path }),
        uppercase: ({ path }) => t('validation.string.uppercase', { path }),
      },
      number: {
        min: ({ path, min }) => t('validation.number.min', { path, min }),
        max: ({ path, max }) => t('validation.number.max', { path, max }),
        lessThan: ({ path, less }) => t('validation.number.lessThan', { path, less }),
        moreThan: ({ path, more }) => t('validation.number.moreThan', { path, more }),
        positive: ({ path }) => t('validation.number.positive', { path }),
        negative: ({ path }) => t('validation.number.negative', { path }),
        integer: ({ path }) => t('validation.number.integer', { path }),
      },
      date: {
        min: ({ path, min }) => t('validation.date.min', { path, min }),
        max: ({ path, max }) => t('validation.date.max', { path, max }),
      },
      object: {
        noUnknown: ({ path }) => t('validation.object.noUnknown', { path }),
      },
      array: {
        min: ({ min }) => t('validation.array.min', { min }),
        max: ({ max }) => t('validation.array.max', { max }),
      },
    })
  }

  watch(locale, apply, { immediate: true })
})
