// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { array, number, object, string } from 'yup'
import { getFieldMeta, validateValues } from './schema'

const schema = object({
  email: string().required().label('Email'),
  age: number().optional(),
  address: object({ city: string().required() }),
  tags: array().of(string().required()).required(),
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
    expect(Object.keys(errors).toSorted()).toEqual(['address.city', 'email'])
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
