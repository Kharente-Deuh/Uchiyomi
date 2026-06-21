// SPDX-License-Identifier: AGPL-3.0-or-later
import type { Ref } from 'vue'
import { watch } from 'vue'
import { setLocale } from 'yup'
import { fr } from 'yup-locales'

/** Minimal shape of an i18n translate function (vue-i18n's `t`). */
export type Translate = (key: string, named?: Record<string, unknown>) => string

/**
 * Wire yup's global validation messages to the app i18n (`validation.*` keys)
 * and re-apply them whenever the active locale changes.
 *
 * Extracted from the Nuxt plugin so the wiring is unit-testable without a Nuxt
 * runtime. `setLocale` mutates yup's global state, so this must only run on the
 * client (the plugin that calls it is `.client`-suffixed).
 */
export function setupYupLocale(t: Translate, locale: Ref<string>): void {
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
}
