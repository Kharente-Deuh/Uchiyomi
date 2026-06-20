// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import en from '../i18n/locales/en.json'
import fr from '../i18n/locales/fr.json'

function keys(obj: Record<string, unknown>, prefix = ''): string[] {
  return Object.entries(obj).flatMap(([k, v]) => {
    const path = prefix ? `${prefix}.${k}` : k

    return v && typeof v === 'object' ? keys(v as Record<string, unknown>, path) : [path]
  })
}

describe('i18n locales', () => {
  it('en and fr expose the same keys', () => {
    expect(keys(fr).toSorted()).toEqual(keys(en).toSorted())
  })
})
