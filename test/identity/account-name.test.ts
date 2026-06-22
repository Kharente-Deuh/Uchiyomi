// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { accountNameSchema } from '../../server/utils/account-name'
import { normalizeAccountName } from '../../shared/dto/identity/account-name'

describe('normalizeAccountName', () => {
  it('trims and lowercases', () => {
    expect(normalizeAccountName('  Admin ')).toBe('admin')
  })
})

describe('accountNameSchema', () => {
  it('accepts a valid name', () => {
    expect(accountNameSchema.parse('alice_01')).toBe('alice_01')
  })

  it('normalizes Admin → admin', () => {
    expect(accountNameSchema.parse('Admin')).toBe('admin')
  })

  it('rejects too short (< 3)', () => {
    expect(accountNameSchema.safeParse('ab').success).toBe(false)
  })

  it('rejects too long (> 32)', () => {
    expect(accountNameSchema.safeParse('a'.repeat(33)).success).toBe(false)
  })

  it('rejects illegal characters', () => {
    expect(accountNameSchema.safeParse('has space').success).toBe(false)
    expect(accountNameSchema.safeParse('a@b.c').success).toBe(false)
  })
})
