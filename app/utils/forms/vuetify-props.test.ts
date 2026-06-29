// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { nextTick, ref } from 'vue'
import { object, string } from 'yup'
import { useForm } from './use-form'

const flushPromises = (): Promise<void> => new Promise(resolve => setTimeout(resolve, 0))

const schema = object({ title: string().required('required').label('Titre') })

describe('vuetify field props', () => {
  it('exposes vuetify-shaped props on field().props', () => {
    const form = useForm({ schema, initialValues: { title: 'hi' } })
    const props = form.field('title').props
    expect(props.modelValue).toBe('hi')
    expect(props.label).toBe('Titre')
    expect(props.required).toBe(true)
  })

  it('exposes errors only once touched', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    const props = form.field('title').props
    await nextTick()
    expect(props.errorMessages).toEqual([])
    form.field('title').handleBlur()
    await nextTick()
    expect(props.errorMessages).toEqual(['required'])
  })

  it('updates value through onUpdate:modelValue', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    form.field('title').props['onUpdate:modelValue']('typed')
    await nextTick()
    expect(form.values.value.title).toBe('typed')
  })

  it('reflects form-level disabled on field props (reactive)', () => {
    const disabled = ref(false)
    const form = useForm({ schema, initialValues: { title: '' }, disabled })
    expect(form.field('title').props.disabled).toBe(false)
    disabled.value = true
    expect(form.field('title').props.disabled).toBe(true)
  })
})

describe('field-required / field-valid classes', () => {
  it('applies field-required on a required field', () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    expect(form.field('title').props.class).toContain('field-required')
  })

  it('does not apply field-valid while a required field is invalid', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    await flushPromises()
    expect(form.field('title').props.class).toContain('field-required')
    expect(form.field('title').props.class).not.toContain('field-valid')
  })

  it('applies field-valid once a required field becomes valid', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    form.field('title').handleChange('ok')
    await flushPromises()
    const cls = form.field('title').props.class
    expect(cls).toContain('field-required')
    expect(cls).toContain('field-valid')
  })

  it('applies no field-* classes on an optional field', async () => {
    const optionalSchema = object({ note: string().optional() })
    const form = useForm({ schema: optionalSchema, initialValues: { note: '' } })
    await flushPromises()
    expect(form.field('note').props.class).toEqual([])
  })
})
