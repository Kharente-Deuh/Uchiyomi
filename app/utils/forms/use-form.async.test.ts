// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { object, string } from 'yup'
import { useForm } from './use-form'

/**
 * Tests for:
 *   - async race guard (stale validation results are dropped)
 *   - validateOn modes (blur / change / submit)
 */

const flushPromises = (): Promise<void> => new Promise(resolve => setTimeout(resolve, 0))

// ---------------------------------------------------------------------------
// Async race guard
// ---------------------------------------------------------------------------

describe('async race guard', () => {
  it('drops stale validation results when a newer run completes first', async () => {
    let resolveSlow!: () => void
    let resolveFast!: () => void
    const slowPromise = new Promise<void>((res) => {
      resolveSlow = res
    })
    const fastPromise = new Promise<void>((res) => {
      resolveFast = res
    })

    const asyncSchema = object({
      name: string().test('ordered', 'invalid', async (v) => {
        if (v === 'slow') {
          await slowPromise

          return false
        }

        if (v === 'fast') {
          await fastPromise

          return true
        }

        return true
      }),
    })

    const form = useForm({ schema: asyncSchema, initialValues: { name: '' } })

    form.field('name').handleChange('slow')
    await flushPromises()

    form.field('name').handleChange('fast')

    resolveFast()
    await flushPromises()
    expect(form.isValid.value).toBe(true)

    resolveSlow()
    await flushPromises()
    expect(form.isValid.value).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// validateOn modes
// ---------------------------------------------------------------------------

const simpleSchema = object({ title: string().required('required') })

describe('validateOn: blur (default)', () => {
  it('shows errors only after handleBlur', async () => {
    const form = useForm({ schema: simpleSchema, initialValues: { title: '' } })
    await flushPromises()
    expect(form.field('title').errors.value).toEqual([])
    form.field('title').handleChange('')
    expect(form.field('title').errors.value).toEqual([])
    form.field('title').handleBlur()
    expect(form.field('title').errors.value).toEqual(['required'])
  })
})

describe('validateOn: change', () => {
  it('shows errors after first handleChange (not waiting for blur)', async () => {
    const form = useForm({ schema: simpleSchema, initialValues: { title: '' }, validateOn: 'change' })
    await flushPromises()
    expect(form.field('title').errors.value).toEqual([])
    form.field('title').handleChange('')
    await flushPromises()
    expect(form.field('title').errors.value).toEqual(['required'])
  })
})

describe('validateOn: submit', () => {
  it('errors are hidden until handleSubmit — neither handleChange nor handleBlur reveals them', async () => {
    const form = useForm({ schema: simpleSchema, initialValues: { title: '' }, validateOn: 'submit', scrollToError: false })
    await flushPromises()
    form.field('title').handleChange('')
    expect(form.field('title').errors.value).toEqual([])
    form.field('title').handleBlur()
    expect(form.field('title').errors.value).toEqual([])
    await form.handleSubmit()
    expect(form.field('title').errors.value).toEqual(['required'])
  })

  it('isValid is still accurate (validation runs regardless of validateOn)', async () => {
    const form = useForm({ schema: simpleSchema, initialValues: { title: '' }, validateOn: 'submit' })
    await flushPromises()
    expect(form.isValid.value).toBe(false)
    form.field('title').handleChange('ok')
    await flushPromises()
    expect(form.isValid.value).toBe(true)
  })
})
