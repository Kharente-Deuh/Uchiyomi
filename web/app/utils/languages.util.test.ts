// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { endomyLanguage } from './languages.util'

describe('endomyLanguage', () => {
  it('returns the endonym for English ("en")', () => {
    expect(endomyLanguage('en')).toBe('English')
  })

  it('returns the endonym for French ("fr")', () => {
    expect(endomyLanguage('fr')).toBe('français')
  })

  it('returns the endonym for Spanish ("es")', () => {
    expect(endomyLanguage('es')).toBe('español')
  })

  it('returns the endonym for Japanese ("ja")', () => {
    expect(endomyLanguage('ja')).toBe('日本語')
  })

  it('returns the endonym for Chinese simplified ("zh")', () => {
    const result = endomyLanguage('zh')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })

  it('returns the endonym for Korean ("ko")', () => {
    expect(endomyLanguage('ko')).toBe('한국어')
  })

  it('returns the endonym for German ("de")', () => {
    expect(endomyLanguage('de')).toBe('Deutsch')
  })

  it('returns the endonym for Portuguese ("pt")', () => {
    const result = endomyLanguage('pt')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })

  it('always returns a string (even for unusual but valid BCP-47 tags)', () => {
    const result = endomyLanguage('all')
    expect(typeof result).toBe('string')
  })

  it('casts the return value to string (non-undefined)', () => {
    const result = endomyLanguage('en')
    expect(result).not.toBeUndefined()
  })
})
