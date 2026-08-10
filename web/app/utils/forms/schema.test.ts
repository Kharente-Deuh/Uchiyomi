// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { array, boolean, number, object, string } from 'yup'
import { getFieldMeta, validateValues } from './schema'

const schema = object({
  email: string().required().label('Email'),
  age: number().optional(),
  address: object({ city: string().required() }),
  tags: array().of(string().required()).required(),
})

const definedSchema = object({
  claim: string().defined(),
  values: array().of(string().required()).defined(),
  bounded: string().defined().min(3),
  flag: boolean().required(),
  scopes: array().of(string().required()).required().min(1),
})

describe('getFieldMeta', () => {
  it('reads label and required from a required field', () => {
    expect(getFieldMeta(schema, 'email')).toEqual({ label: 'Email', required: true })
  })

  it('marks optional fields as not required', () => {
    expect(getFieldMeta(schema, 'age')).toEqual({ label: undefined, required: false })
  })

  it('reads nested field meta', () => {
    expect(getFieldMeta(schema, 'address.city')).toEqual({ label: undefined, required: true })
  })

  it('does not mark a defined string as required', () => {
    expect(getFieldMeta(definedSchema, 'claim').required).toBe(false)
  })

  it('does not mark a defined array as required', () => {
    expect(getFieldMeta(definedSchema, 'values').required).toBe(false)
  })

  it('marks a defined field whose empty value fails a test as required', () => {
    expect(getFieldMeta(definedSchema, 'bounded').required).toBe(true)
  })

  it('marks a required boolean as required', () => {
    expect(getFieldMeta(definedSchema, 'flag').required).toBe(true)
  })

  it('marks a non-empty array as required', () => {
    expect(getFieldMeta(definedSchema, 'scopes').required).toBe(true)
  })
})

describe('validateValues', () => {
  it('returns empty object for valid values', async () => {
    const errors = await validateValues(schema, {
      email: 'a@b.com',
      address: { city: 'Paris' },
      tags: ['x'],
    })
    expect(errors).toEqual({})
  })

  it('maps each invalid path to its first message', async () => {
    const errors = await validateValues(schema, {
      email: '',
      address: { city: '' },
      tags: ['x'],
    })
    expect(Object.keys(errors).toSorted((a, b) => a.localeCompare(b))).toEqual(['address.city', 'email'])
    expect(errors.email).toHaveLength(1)
  })

  it('supports async tests', async () => {
    const asyncSchema = object({
      name: string().test('async-ok', 'too short', async v => (v?.length ?? 0) > 2),
    })
    const errors = await validateValues(asyncSchema, { name: 'a' })
    expect(errors.name).toEqual(['too short'])
  })
})
