// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import * as yup from 'yup'
import { setupYupLocale } from './yup-locale'

function tStub(key: string): string {
  return key
}

const setLocaleSpy = vi.spyOn(yup, 'setLocale').mockImplementation(() => {})

describe('setupYupLocale', () => {
  beforeEach(() => {
    setLocaleSpy.mockClear()
  })

  it('calls setLocale immediately (immediate: true) when locale is "en"', () => {
    const locale = ref('en')
    const t = tStub

    setupYupLocale(t, locale)

    expect(setLocaleSpy).toHaveBeenCalledTimes(1)
  })

  it('calls setLocale twice on setup when locale is "fr" (fr base + custom)', () => {
    const locale = ref('fr')
    const t = tStub

    setupYupLocale(t, locale)

    expect(setLocaleSpy).toHaveBeenCalledTimes(2)
  })

  it('re-applies setLocale when locale changes from "en" to "fr"', async () => {
    const locale = ref('en')
    const t = tStub

    setupYupLocale(t, locale)
    setLocaleSpy.mockClear()

    locale.value = 'fr'
    await Promise.resolve()

    expect(setLocaleSpy).toHaveBeenCalledTimes(2)
  })

  it('re-applies setLocale when locale changes from "fr" to "en"', async () => {
    const locale = ref('fr')
    const t = tStub

    setupYupLocale(t, locale)
    setLocaleSpy.mockClear()

    locale.value = 'en'
    await Promise.resolve()

    expect(setLocaleSpy).toHaveBeenCalledTimes(1)
  })

  it('passes the "validation.mixed.required" i18n key to setLocale for mixed.required', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

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
      mixed?: { oneOf?: (parameters: { path: string, values: string }) => string }
    }
    const function_ = customLocale.mixed?.oneOf
    expect(typeof function_).toBe('function')
    function_!({ path: 'field', values: 'a,b' })
    expect(t).toHaveBeenCalledWith('validation.mixed.oneOf', { path: 'field', values: 'a,b' })
  })

  it('passes a function for mixed.notOneOf that calls t with "validation.mixed.notOneOf"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      mixed?: { notOneOf?: (parameters: { path: string, values: string }) => string }
    }
    const function_ = customLocale.mixed?.notOneOf
    function_!({ path: 'field', values: 'a,b' })
    expect(t).toHaveBeenCalledWith('validation.mixed.notOneOf', { path: 'field', values: 'a,b' })
  })

  it('passes a function for string.min that calls t with "validation.string.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { min?: (parameters: { path: string, min: number }) => string }
    }
    const function_ = customLocale.string?.min
    function_!({ path: 'username', min: 3 })
    expect(t).toHaveBeenCalledWith('validation.string.min', { path: 'username', min: 3 })
  })

  it('passes a function for string.max that calls t with "validation.string.max"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { max?: (parameters: { path: string, max: number }) => string }
    }
    const function_ = customLocale.string?.max
    function_!({ path: 'username', max: 32 })
    expect(t).toHaveBeenCalledWith('validation.string.max', { path: 'username', max: 32 })
  })

  it('passes a function for string.matches that calls t with "validation.string.matches"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { matches?: (parameters: { path: string, regex: RegExp }) => string }
    }
    const function_ = customLocale.string?.matches
    function_!({ path: 'username', regex: /^[a-z]+$/ })
    expect(t).toHaveBeenCalledWith('validation.string.matches', { path: 'username', regex: /^[a-z]+$/ })
  })

  it('passes a function for string.email that calls t with "validation.string.email"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      string?: { email?: (parameters: { path: string }) => string }
    }
    customLocale.string?.email!({ path: 'email' })
    expect(t).toHaveBeenCalledWith('validation.string.email', { path: 'email' })
  })

  it('passes a function for number.min that calls t with "validation.number.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      number?: { min?: (parameters: { path: string, min: number }) => string }
    }
    customLocale.number?.min!({ path: 'count', min: 1 })
    expect(t).toHaveBeenCalledWith('validation.number.min', { path: 'count', min: 1 })
  })

  it('passes a function for number.integer that calls t with "validation.number.integer"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      number?: { integer?: (parameters: { path: string }) => string }
    }
    customLocale.number?.integer!({ path: 'count' })
    expect(t).toHaveBeenCalledWith('validation.number.integer', { path: 'count' })
  })

  it('passes a function for array.min that calls t with "validation.array.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      array?: { min?: (parameters: { min: number }) => string }
    }
    customLocale.array?.min!({ min: 1 })
    expect(t).toHaveBeenCalledWith('validation.array.min', { min: 1 })
  })

  it('passes a function for object.noUnknown that calls t with "validation.object.noUnknown"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      object?: { noUnknown?: (parameters: { path: string }) => string }
    }
    customLocale.object?.noUnknown!({ path: 'obj' })
    expect(t).toHaveBeenCalledWith('validation.object.noUnknown', { path: 'obj' })
  })

  it('passes a function for date.min that calls t with "validation.date.min"', () => {
    const locale = ref('en')
    const t = vi.fn((key: string) => key)

    setupYupLocale(t, locale)

    const customLocale = setLocaleSpy.mock.calls[0]![0] as {
      date?: { min?: (parameters: { path: string, min: unknown }) => string }
    }
    customLocale.date?.min!({ path: 'dob', min: '2000-01-01' })
    expect(t).toHaveBeenCalledWith('validation.date.min', { path: 'dob', min: '2000-01-01' })
  })
})
