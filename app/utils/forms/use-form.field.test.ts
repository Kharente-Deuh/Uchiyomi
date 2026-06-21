// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { object, string } from 'yup'
import { useForm } from './use-form'

const flushPromises = (): Promise<void> => new Promise(resolve => setTimeout(resolve, 0))

const schema = object({
  title: string().required('title required').label('Titre'),
  desc: string().optional(),
})

describe('useForm — field & validation', () => {
  it('exposes initial values', () => {
    const form = useForm({ schema, initialValues: { title: 'hi', desc: '' } })
    expect(form.values.value).toEqual({ title: 'hi', desc: '' })
    expect(form.field('title').value.value).toBe('hi')
  })

  it('derives label and required from the schema', () => {
    const form = useForm({ schema, initialValues: { title: '', desc: '' } })
    expect(form.field('title').label).toBe('Titre')
    expect(form.field('title').required).toBe(true)
    expect(form.field('desc').required).toBe(false)
  })

  it('memoizes field() per path', () => {
    const form = useForm({ schema, initialValues: { title: '', desc: '' } })
    expect(form.field('title')).toBe(form.field('title'))
  })

  it('updates value through handleChange', async () => {
    const form = useForm({ schema, initialValues: { title: '', desc: '' } })
    form.field('title').handleChange('new')
    await nextTick()
    expect(form.values.value.title).toBe('new')
  })

  it('computes global validity eagerly, regardless of touched', async () => {
    const form = useForm({ schema, initialValues: { title: '', desc: '' } })
    await flushPromises()
    expect(form.isValid.value).toBe(false)
    form.field('title').handleChange('ok')
    await flushPromises()
    expect(form.isValid.value).toBe(true)
  })

  it('isValid handles async validators correctly', async () => {
    const asyncSchema = object({
      name: string().test('async-check', 'fail', v => Promise.resolve(v !== 'bad')),
    })
    const form = useForm({ schema: asyncSchema, initialValues: { name: 'bad' } })
    await flushPromises()
    expect(form.isValid.value).toBe(false)
    form.field('name').handleChange('good')
    await flushPromises()
    expect(form.isValid.value).toBe(true)
  })

  it('hides field errors until the field is touched, then shows them', async () => {
    const form = useForm({ schema, initialValues: { title: '', desc: '' } })
    await nextTick()
    expect(form.field('title').errors.value).toEqual([])
    form.field('title').handleBlur()
    await nextTick()
    expect(form.field('title').errors.value).toEqual(['title required'])
  })
})
