import { beforeEach, describe, expect, it, vi } from 'vitest'
// SPDX-License-Identifier: AGPL-3.0-or-later
import { ref } from 'vue'
import * as yup from 'yup'
import { setupYupLocale } from './yup-locale'

function tStub(key: string): string {
  return key
}

// Spy on yup.setLocale so we can capture the locale objects passed to it
// without actually mutating yup's global state (which would bleed between tests).
const setLocaleSpy = vi.spyOn(yup, 'setLocale').mockImplementation(() => {})

describe('setupYupLocale', () => {
  // Reset call history between tests
  beforeEach(() => {
    setLocaleSpy.mockClear()
  })

  // --- immediate invocation on setup ---

  it('calls setLocale immediately (immediate: true) when locale is "en"', () => {
    const locale = ref('en')
    const t = tStub

    setupYupLocale(t, locale)

    // For 'en', only the custom locale object is applied (no fr base)
    expect(setLocaleSpy).toHaveBeenCalledTimes(1)
  })

  it('calls setLocale twice on setup when locale is "fr" (fr base + custom)', () => {
    const locale = ref('fr')
    const t = tStub

    setupYupLocale(t, locale)

    // First call is the yup-locales fr base, second is the custom overrides
    expect(setLocaleSpy).toHaveBeenCalledTimes(2)
  })

  // --- locale change re-applies ---

  it('re-applies setLocale when locale changes from "en" to "fr"', async () => {
    const locale = ref('en')
    const t = tStub

    setupYupLocale(t, locale)
    setLocaleSpy.mockClear()

    locale.value = 'fr'
    // allow Vue's watch flush
    await Promise.resolve()

    expect(setLocaleSpy).toHaveBeenCalledTimes(2) // fr base + custom
  })

  it('re-applies setLocale when locale changes from "fr" to "en"', async () => {
    const locale = ref('fr')
    const t = tStub

    setupYupLocale(t, locale)
    setLocaleSpy.mockClear()

    locale.value = 'en'
    await Promise.resolve()

    expect(setLocaleSpy).toHaveBeenCalledTimes(1) // only custom, no fr base
  })

  // --- i18n key forwarding ---

  it('passes the "validation.mixed.required" i18n key to setLocale for mixed.required', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    // Find the custom locale object call (last call when 'en', only call)
    const customLocale = setLocaleSpy.mock.calls[0]![0] as { mixed?: { required?: string } }
    expect(customLocale.mixed?.required).toBe('validation.mixed.required')
  })

  it('passes the "validation.mixed.default" i18n key to setLocale for mixed.default', () => {
    const locale = ref('en')
    const t = tStub

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as { mixed?: { default?: string } }
    expect(customLocale.mixed?.default).toBe('validation.mixed.default')
  })

  it('passes a function for mixed.oneOf that calls t with "validation.mixed.oneOf"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      mixed?: { oneOf?: (params: { path: string, values: string }) => string }
    }
    const fn = customLocale.mixed?.oneOf
    expect(typeof fn).toBe('function')
    fn!({ path: 'field', values: 'a,b' })
    expect(t).toHaveBeenCalledWith('validation.mixed.oneOf', { path: 'field', values: 'a,b' })
  })

  it('passes a function for mixed.notOneOf that calls t with "validation.mixed.notOneOf"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      mixed?: { notOneOf?: (params: { path: string, values: string }) => string }
    }
    const fn = customLocale.mixed?.notOneOf
    fn!({ path: 'field', values: 'a,b' })
    expect(t).toHaveBeenCalledWith('validation.mixed.notOneOf', { path: 'field', values: 'a,b' })
  })

  // --- string messages ---

  it('passes a function for string.min that calls t with "validation.string.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { min?: (params: { path: string, min: number }) => string }
    }
    const fn = customLocale.string?.min
    fn!({ path: 'username', min: 3 })
    expect(t).toHaveBeenCalledWith('validation.string.min', { path: 'username', min: 3 })
  })

  it('passes a function for string.max that calls t with "validation.string.max"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { max?: (params: { path: string, max: number }) => string }
    }
    const fn = customLocale.string?.max
    fn!({ path: 'username', max: 32 })
    expect(t).toHaveBeenCalledWith('validation.string.max', { path: 'username', max: 32 })
  })

  it('passes a function for string.matches that calls t with "validation.string.matches"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { matches?: (params: { path: string, regex: RegExp }) => string }
    }
    const fn = customLocale.string?.matches
    fn!({ path: 'username', regex: /^[a-z]+$/ })
    expect(t).toHaveBeenCalledWith('validation.string.matches', { path: 'username', regex: /^[a-z]+$/ })
  })

  it('passes a function for string.email that calls t with "validation.string.email"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { email?: (params: { path: string }) => string }
    }
    customLocale.string?.email!({ path: 'email' })
    expect(t).toHaveBeenCalledWith('validation.string.email', { path: 'email' })
  })

  // --- number messages ---

  it('passes a function for number.min that calls t with "validation.number.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      number?: { min?: (params: { path: string, min: number }) => string }
    }
    customLocale.number?.min!({ path: 'count', min: 1 })
    expect(t).toHaveBeenCalledWith('validation.number.min', { path: 'count', min: 1 })
  })

  it('passes a function for number.integer that calls t with "validation.number.integer"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      number?: { integer?: (params: { path: string }) => string }
    }
    customLocale.number?.integer!({ path: 'count' })
    expect(t).toHaveBeenCalledWith('validation.number.integer', { path: 'count' })
  })

  // --- array messages ---

  it('passes a function for array.min that calls t with "validation.array.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      array?: { min?: (params: { min: number }) => string }
    }
    customLocale.array?.min!({ min: 1 })
    expect(t).toHaveBeenCalledWith('validation.array.min', { min: 1 })
  })

  // --- object messages ---

  it('passes a function for object.noUnknown that calls t with "validation.object.noUnknown"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      object?: { noUnknown?: (params: { path: string }) => string }
    }
    customLocale.object?.noUnknown!({ path: 'obj' })
    expect(t).toHaveBeenCalledWith('validation.object.noUnknown', { path: 'obj' })
  })

  // --- date messages ---

  it('passes a function for date.min that calls t with "validation.date.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      date?: { min?: (params: { path: string, min: unknown }) => string }
    }
    customLocale.date?.min!({ path: 'dob', min: '2000-01-01' })
    expect(t).toHaveBeenCalledWith('validation.date.min', { path: 'dob', min: '2000-01-01' })
  })
})
