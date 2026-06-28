// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { object, string } from 'yup'
import { useForm } from './use-form'

const schema = object({ title: string().required('required') })

describe('handleSubmit', () => {
  it('calls onSubmit with typed values when valid', async () => {
    const onSubmit = vi.fn()
    const form = useForm({ schema, initialValues: { title: 'ok' }, onSubmit })
    await form.handleSubmit()
    expect(onSubmit).toHaveBeenCalledWith({ title: 'ok' })
  })

  it('does not call onSubmit when invalid and reveals errors', async () => {
    const onSubmit = vi.fn()
    const form = useForm({ schema, initialValues: { title: '' }, onSubmit, scrollToError: false })
    await form.handleSubmit()
    expect(onSubmit).not.toHaveBeenCalled()
    expect(form.field('title').errors.value).toEqual(['required'])
  })
})

describe('setServerErrors', () => {
  it('sets errors per path and reveals them', async () => {
    const form = useForm({ schema, initialValues: { title: 'taken' }, scrollToError: false })
    form.setServerErrors({ title: 'already used' })
    expect(form.field('title').errors.value).toEqual(['already used'])
  })

  it('clears a server error when the field changes', async () => {
    const form = useForm({ schema, initialValues: { title: 'taken' }, scrollToError: false })
    form.setServerErrors({ title: 'already used' })
    expect(form.field('title').errors.value).toEqual(['already used'])
    form.field('title').handleChange('fresh')
    await nextTick()
    expect(form.field('title').errors.value).toEqual([])
  })
})
