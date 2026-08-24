// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { slugFromDisplayName } from './slug-from-display-name'

describe('slugFromDisplayName', () => {
  it('slugifies a simple name', () => {
    expect(slugFromDisplayName('Acme SSO')).toBe('acme-sso')
  })

  it('returns empty when nothing valid remains', () => {
    expect(slugFromDisplayName('日本語')).toBe('')
  })

  it('returns empty when longer than 64', () => {
    expect(slugFromDisplayName('A'.repeat(65))).toBe('')
  })
})
