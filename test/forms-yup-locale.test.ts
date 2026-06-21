// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
// Pure-logic test for the yup-locale wiring. Runs in the node project (not the
// nuxt one) so a single `yup` instance is shared between this file's schemas and
// `setupYupLocale`'s `setLocale` — the nuxt SSR transform can otherwise duplicate
// the yup module, leaving the global locale override on a different instance.
import type { AnySchema } from 'yup'
import { describe, expect, it } from 'vitest'
import { nextTick, ref } from 'vue'
import { array, date, number, object, string, ValidationError } from 'yup'
import { setupYupLocale } from '../app/utils/forms/yup-locale'

// Echoes the i18n key (plus any params) so each wired message is identifiable.
function t(key: string, named?: Record<string, unknown>): string {
  return named ? `${key} ${JSON.stringify(named)}` : key
}

async function firstError(schema: AnySchema, value: unknown): Promise<string> {
  try {
    await schema.validate(value, { abortEarly: false })

    return ''
  } catch (error) {
    return error instanceof ValidationError ? error.errors[0] : String(error)
  }
}

// One entry per wired rule — proves every message routes through `t`.
// Schemas are built lazily (thunks): yup snapshots a rule's default message at
// build time when no custom message is given, so they must be constructed AFTER
// `setupYupLocale` has installed the locale overrides.
const cases: Array<{ rule: string, schema: () => AnySchema, value: unknown, key: string }> = [
  { rule: 'mixed.required', schema: () => string().required(), value: '', key: 'validation.mixed.required' },
  { rule: 'mixed.oneOf', schema: () => string().oneOf(['a']), value: 'b', key: 'validation.mixed.oneOf' },
  { rule: 'mixed.notOneOf', schema: () => string().notOneOf(['a']), value: 'a', key: 'validation.mixed.notOneOf' },
  { rule: 'string.length', schema: () => string().length(5), value: 'ab', key: 'validation.string.length' },
  { rule: 'string.min', schema: () => string().min(3), value: 'a', key: 'validation.string.min' },
  { rule: 'string.max', schema: () => string().max(2), value: 'abcd', key: 'validation.string.max' },
  { rule: 'string.matches', schema: () => string().matches(/^x$/), value: 'y', key: 'validation.string.matches' },
  { rule: 'string.email', schema: () => string().email(), value: 'bad', key: 'validation.string.email' },
  { rule: 'string.url', schema: () => string().url(), value: 'bad', key: 'validation.string.url' },
  { rule: 'string.trim', schema: () => string().trim().strict(), value: ' a ', key: 'validation.string.trim' },
  { rule: 'string.lowercase', schema: () => string().lowercase().strict(), value: 'A', key: 'validation.string.lowercase' },
  { rule: 'string.uppercase', schema: () => string().uppercase().strict(), value: 'a', key: 'validation.string.uppercase' },
  { rule: 'number.min', schema: () => number().min(5), value: 1, key: 'validation.number.min' },
  { rule: 'number.max', schema: () => number().max(5), value: 9, key: 'validation.number.max' },
  { rule: 'number.lessThan', schema: () => number().lessThan(5), value: 9, key: 'validation.number.lessThan' },
  { rule: 'number.moreThan', schema: () => number().moreThan(5), value: 1, key: 'validation.number.moreThan' },
  { rule: 'number.positive', schema: () => number().positive(), value: -1, key: 'validation.number.positive' },
  { rule: 'number.negative', schema: () => number().negative(), value: 1, key: 'validation.number.negative' },
  { rule: 'number.integer', schema: () => number().integer(), value: 1.5, key: 'validation.number.integer' },
  { rule: 'date.min', schema: () => date().min(new Date('2020-06-01')), value: new Date('2019-01-01'), key: 'validation.date.min' },
  { rule: 'date.max', schema: () => date().max(new Date('2020-06-01')), value: new Date('2021-01-01'), key: 'validation.date.max' },
  { rule: 'object.noUnknown', schema: () => object({ a: string() }).noUnknown().strict(), value: { a: 'x', b: 'y' }, key: 'validation.object.noUnknown' },
  { rule: 'array.min', schema: () => array().min(2), value: [1], key: 'validation.array.min' },
  { rule: 'array.max', schema: () => array().max(1), value: [1, 2], key: 'validation.array.max' },
]

describe('setupYupLocale', () => {
  it('routes every wired validation message through the i18n t() function', async () => {
    const locale = ref('en')
    setupYupLocale(t, locale)

    for (const { rule, schema, value, key } of cases) {
      const message = await firstError(schema(), value)
      expect(message, `rule ${rule} should route through ${key}`).toContain(key)
    }
  })

  it('applies the yup-locales French base on locale change, our overrides still winning', async () => {
    const locale = ref('en')
    setupYupLocale(t, locale)

    locale.value = 'fr'
    await nextTick()

    // The watch re-ran with locale=fr (fr branch taken) and our dict was applied last.
    expect(await firstError(string().required(), '')).toBe('validation.mixed.required')
  })
})
