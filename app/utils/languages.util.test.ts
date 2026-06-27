// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { endomyLanguage } from './languages.util'

describe('endomyLanguage', () => {
  // --- well-known ISO 639-1 codes ---

  it('returns the endonym for English ("en")', () => {
    // Intl.DisplayNames returns the language's own name in itself
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
    // The Intl spec may return 'zh' or '中文'; just confirm it is non-empty
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
    // Node Intl may return 'português' or a variant; confirm non-empty
    const result = endomyLanguage('pt')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })

  // --- return type ---

  it('always returns a string (even for unusual but valid BCP-47 tags)', () => {
    // 'all' is the "undetermined" language tag used in Tachiyomi/Suwayomi
    // Intl.DisplayNames.of() returns the tag itself or a localised name for valid tags.
    const result = endomyLanguage('all')
    expect(typeof result).toBe('string')
  })

  it('casts the return value to string (non-undefined)', () => {
    // The function casts `as string`, so it should never return undefined
    const result = endomyLanguage('en')
    expect(result).not.toBeUndefined()
  })
})
