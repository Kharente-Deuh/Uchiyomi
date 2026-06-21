// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { object, string } from 'yup'
import { useForm } from './use-form'

describe('useForm — edge coverage', () => {
  it('builds missing parent objects when setting a nested path', async () => {
    const schema = object({ nested: object({ value: string().optional() }) })
    const form = useForm({ schema, initialValues: {} as any })
    form.field('nested.value').handleChange('x')
    await nextTick()
    expect(form.values.value).toEqual({ nested: { value: 'x' } })
    // The nested leaf participates in dirty tracking (flatten recursion).
    expect(form.isDirty.value).toBe(true)
  })

  it('debounces validation when asyncDebounceMs is set', async () => {
    const schema = object({ title: string().required('required') })
    const form = useForm({ schema, initialValues: { title: '' }, asyncDebounceMs: 5 })
    form.field('title').handleChange('')
    form.field('title').handleChange('') // second change clears the prior timer
    await new Promise(resolve => setTimeout(resolve, 20))
    expect(form.isValid.value).toBe(false)
    form.field('title').handleChange('ok')
    await new Promise(resolve => setTimeout(resolve, 20))
    expect(form.isValid.value).toBe(true)
  })

  it('exposes per-field touched state', async () => {
    const schema = object({ title: string().required() })
    const form = useForm({ schema, initialValues: { title: '' } })
    const field = form.field('title')
    expect(field.isTouched.value).toBe(false)
    field.handleBlur()
    await nextTick()
    expect(field.isTouched.value).toBe(true)
  })

  it('runs the scroll-to-error path on invalid submit (default scrollToError)', async () => {
    const schema = object({ title: string().required('required') })
    const form = useForm({ schema, initialValues: { title: '' } })
    await form.handleSubmit()
    expect(form.field('title').errors.value).toEqual(['required'])
  })
})
