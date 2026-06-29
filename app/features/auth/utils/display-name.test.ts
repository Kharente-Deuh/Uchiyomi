// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import {
  DISPLAY_NAME_MAX,
  DISPLAY_NAME_MIN,
} from '#shared/dto/identity/display-name'
import { displayNameRule } from './display-name'

describe('displayNameRule', () => {
  const rule = displayNameRule('Display name')

  // --- label ---

  it('carries the label supplied by the caller', () => {
    const schema = displayNameRule('My Label')
    expect((schema as unknown as { spec: { label: string } }).spec.label).toBe('My Label')
  })

  // --- valid values ---

  it('accepts a plain name', async () => {
    await expect(rule.isValid('Alice')).resolves.toBe(true)
  })

  it('accepts a name with spaces and mixed case', async () => {
    await expect(rule.isValid('Alice Bob')).resolves.toBe(true)
  })

  it('accepts special characters (not restricted by pattern)', async () => {
    await expect(rule.isValid('Ñoño 42!')).resolves.toBe(true)
  })

  it('accepts a name of exactly DISPLAY_NAME_MIN characters after trim', async () => {
    // DISPLAY_NAME_MIN = 1; a single non-whitespace character satisfies required+min
    const value = 'a'.repeat(DISPLAY_NAME_MIN)
    await expect(rule.isValid(value)).resolves.toBe(true)
  })

  it('accepts a name of exactly DISPLAY_NAME_MAX characters', async () => {
    const value = 'a'.repeat(DISPLAY_NAME_MAX)
    await expect(rule.isValid(value)).resolves.toBe(true)
  })

  it('strips surrounding whitespace before checking min length', async () => {
    // '  a  ' → trims to 'a' (length 1) — still passes DISPLAY_NAME_MIN = 1
    await expect(rule.isValid('  a  ')).resolves.toBe(true)
  })

  // --- invalid values ---

  it('rejects a string longer than DISPLAY_NAME_MAX', async () => {
    const value = 'a'.repeat(DISPLAY_NAME_MAX + 1)
    await expect(rule.isValid(value)).resolves.toBe(false)
  })

  it('rejects an empty string (required)', async () => {
    await expect(rule.isValid('')).resolves.toBe(false)
  })

  it('rejects a whitespace-only string (trims to empty, then required fails)', async () => {
    await expect(rule.isValid('   ')).resolves.toBe(false)
  })

  it('rejects undefined (required)', async () => {
    // eslint-disable-next-line unicorn/no-useless-undefined
    await expect(rule.isValid(undefined)).resolves.toBe(false)
  })
})
