// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { isReactive, nextTick, reactive, toRaw } from 'vue'
import { array, object, string } from 'yup'
import { useForm } from './use-form'

const schema = object({ a: string().optional(), b: string().optional() })

describe('isDirty & reset', () => {
  it('is false initially, true after a change, false again after reset', async () => {
    const form = useForm({ schema, initialValues: { a: '1', b: '2' } })
    expect(form.isDirty.value).toBe(false)
    form.field('a').handleChange('x')
    await nextTick()
    expect(form.isDirty.value).toBe(true)
    form.reset()
    await nextTick()
    expect(form.isDirty.value).toBe(false)
    expect(form.values.value).toEqual({ a: '1', b: '2' })
  })

  it('ignores key order in dirty comparison', async () => {
    const form = useForm({ schema, initialValues: { a: '1', b: '2' } })
    form.reset({ b: '2', a: '1' } as any)
    await nextTick()
    expect(form.isDirty.value).toBe(false)
  })

  it('ignores fields listed in ignoreFields', async () => {
    const form = useForm({ schema, initialValues: { a: '1', b: '2' }, ignoreFields: ['a'] })
    form.field('a').handleChange('changed')
    await nextTick()
    expect(form.isDirty.value).toBe(false)
  })

  it('rebases the snapshot when reset is given explicit values', async () => {
    const form = useForm({ schema, initialValues: { a: '1', b: '2' } })
    form.reset({ a: '9', b: '2' })
    await nextTick()
    expect(form.isDirty.value).toBe(false)
    expect(form.values.value.a).toBe('9')
  })

  it('reset(next) accepts a reactive object without throwing', async () => {
    const form = useForm({ schema, initialValues: { a: '1', b: '2' } })
    const next = reactive({ a: '9', b: '2' })
    expect(() => form.reset(next)).not.toThrow()
    await nextTick()
    expect(form.values.value).toEqual({ a: '9', b: '2' })
    expect(form.isDirty.value).toBe(false)
  })

  it('reset(next) accepts a plain object holding reactive values', async () => {
    const listSchema = object({ a: string().optional(), list: array().of(string().required()).optional() })
    const form = useForm({ schema: listSchema, initialValues: { a: '1', list: [] } })
    const source = reactive({ a: '9', list: ['x', 'y'] })
    expect(() => form.reset({ a: source.a, list: source.list })).not.toThrow()
    await nextTick()
    expect(form.values.value).toEqual({ a: '9', list: ['x', 'y'] })
    expect(isReactive(toRaw(form.values.value).list)).toBe(false)
  })
})
