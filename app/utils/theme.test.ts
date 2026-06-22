// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { resolveInitialTheme } from './theme'

const valid = ['light', 'dark']

describe('resolveInitialTheme', () => {
  it('keeps a saved theme that is valid', () => {
    expect(resolveInitialTheme('light', valid, 'dark')).toBe('light')
  })

  it('falls back when no theme is saved', () => {
    expect(resolveInitialTheme(undefined, valid, 'dark')).toBe('dark')
  })

  it('falls back when the saved theme is not a known theme', () => {
    expect(resolveInitialTheme('solarized', valid, 'dark')).toBe('dark')
  })
})
