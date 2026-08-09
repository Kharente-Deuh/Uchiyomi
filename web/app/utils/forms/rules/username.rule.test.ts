// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { usernameRule } from './username.rule'

const rule = usernameRule('Username')

describe('usernameRule', () => {
  it('trims and lowercases the value', async () => {
    await expect(rule.validate('  Alice  ')).resolves.toBe('alice')
  })

  it('accepts digits, underscores and hyphens', async () => {
    await expect(rule.validate('a_b-9')).resolves.toBe('a_b-9')
  })

  it('accepts the shortest allowed username', async () => {
    await expect(rule.validate('abc')).resolves.toBe('abc')
  })

  it('accepts the longest allowed username', async () => {
    await expect(rule.validate('a'.repeat(32))).resolves.toBe('a'.repeat(32))
  })

  it('rejects a username shorter than 3 characters', async () => {
    await expect(rule.validate('ab')).rejects.toThrow()
  })

  it('rejects a username longer than 32 characters', async () => {
    await expect(rule.validate('a'.repeat(33))).rejects.toThrow()
  })

  it('rejects characters outside the allowed set', async () => {
    await expect(rule.validate('john.doe')).rejects.toThrow()
  })

  it('rejects a value that is only whitespace', async () => {
    await expect(rule.validate(' '.repeat(4))).rejects.toThrow()
  })

  it('rejects an undefined value', async () => {
    await expect(rule.validate(undefined)).rejects.toThrow()
  })
})
