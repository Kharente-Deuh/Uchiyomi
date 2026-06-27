// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import {
  ACCOUNT_NAME_MAX,
  ACCOUNT_NAME_MIN,
  ACCOUNT_NAME_PATTERN,
} from '#shared/dto/identity/account-name'
import { accountNameRule } from './account-name'

describe('accountNameRule', () => {
  const rule = accountNameRule('Username')

  // --- label ---

  it('carries the label supplied by the caller', () => {
    const schema = accountNameRule('My Label')
    // yup describes() returns the resolved meta; label is on spec
    expect((schema as unknown as { spec: { label: string } }).spec.label).toBe('My Label')
  })

  // --- valid values ---

  it('accepts a lowercase alphanumeric string within bounds', async () => {
    await expect(rule.isValid('alice')).resolves.toBe(true)
  })

  it('accepts an account name with underscores and dashes', async () => {
    await expect(rule.isValid('alice_bob-99')).resolves.toBe(true)
  })

  it('accepts an account name of exactly ACCOUNT_NAME_MIN characters', async () => {
    const value = 'a'.repeat(ACCOUNT_NAME_MIN)
    await expect(rule.isValid(value)).resolves.toBe(true)
  })

  it('accepts an account name of exactly ACCOUNT_NAME_MAX characters', async () => {
    const value = 'a'.repeat(ACCOUNT_NAME_MAX)
    await expect(rule.isValid(value)).resolves.toBe(true)
  })

  // --- lowercase normalization (trim + lowercase transforms) ---

  it('lowercases uppercase input before validating', async () => {
    // ACCOUNT_NAME_PATTERN only allows lowercase; the .lowercase() transform
    // normalises first, so 'Alice' becomes 'alice' and should pass.
    await expect(rule.isValid('Alice')).resolves.toBe(true)
  })

  it('strips leading/trailing whitespace before validating', async () => {
    await expect(rule.isValid('  alice  ')).resolves.toBe(true)
  })

  // --- invalid values ---

  it('rejects a string shorter than ACCOUNT_NAME_MIN', async () => {
    const value = 'a'.repeat(ACCOUNT_NAME_MIN - 1)
    await expect(rule.isValid(value)).resolves.toBe(false)
  })

  it('rejects a string longer than ACCOUNT_NAME_MAX', async () => {
    const value = 'a'.repeat(ACCOUNT_NAME_MAX + 1)
    await expect(rule.isValid(value)).resolves.toBe(false)
  })

  it('rejects a string containing spaces', async () => {
    await expect(rule.isValid('alice bob')).resolves.toBe(false)
  })

  it('rejects a string containing special characters', async () => {
    await expect(rule.isValid('alice!')).resolves.toBe(false)
  })

  it('rejects a string containing dots', async () => {
    await expect(rule.isValid('ali.ce')).resolves.toBe(false)
  })

  it('rejects an empty string', async () => {
    await expect(rule.isValid('')).resolves.toBe(false)
  })

  it('rejects undefined (required)', async () => {
    // eslint-disable-next-line unicorn/no-useless-undefined
    await expect(rule.isValid(undefined)).resolves.toBe(false)
  })

  // --- ACCOUNT_NAME_PATTERN constant integrity ---

  it('aCCOUNT_NAME_PATTERN allows only [a-z0-9_-]', () => {
    expect(ACCOUNT_NAME_PATTERN.test('abc123_-')).toBe(true)
    expect(ACCOUNT_NAME_PATTERN.test('ABC')).toBe(false)
    expect(ACCOUNT_NAME_PATTERN.test('hello world')).toBe(false)
  })
})
