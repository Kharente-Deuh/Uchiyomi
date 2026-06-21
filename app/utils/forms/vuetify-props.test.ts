// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { nextTick, ref } from 'vue'
import { object, string } from 'yup'
import { useForm } from './use-form'

const schema = object({ title: string().required('required').label('Titre') })

describe('vuetify field props', () => {
  it('exposes vuetify-shaped props on field().props', () => {
    const form = useForm({ schema, initialValues: { title: 'hi' } })
    const props = form.field('title').props
    expect(props.value.modelValue).toBe('hi')
    expect(props.value.label).toBe('Titre')
    expect(props.value.required).toBe(true)
  })

  it('exposes errors only once touched', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    const props = form.field('title').props
    await nextTick()
    expect(props.value.errorMessages).toEqual([])
    form.field('title').handleBlur()
    await nextTick()
    expect(props.value.errorMessages).toEqual(['required'])
  })

  it('updates value through onUpdate:modelValue', async () => {
    const form = useForm({ schema, initialValues: { title: '' } })
    form.field('title').props.value['onUpdate:modelValue']('typed')
    await nextTick()
    expect(form.values.value.title).toBe('typed')
  })

  it('reflects form-level disabled on field props (reactive)', () => {
    const disabled = ref(false)
    const form = useForm({ schema, initialValues: { title: '' }, disabled })
    expect(form.field('title').props.value.disabled).toBe(false)
    disabled.value = true
    expect(form.field('title').props.value.disabled).toBe(true)
  })
})
