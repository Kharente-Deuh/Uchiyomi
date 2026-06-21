// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
import { array, object, string } from 'yup'
import { useForm } from './use-form'

const schema = object({
  items: array().of(object({ name: string().required() })).required(),
})

describe('form.array', () => {
  it('reflects initial array items', () => {
    const form = useForm({ schema, initialValues: { items: [{ name: 'a' }, { name: 'b' }] } })
    expect(form.array('items').fields.value.map(f => f.index)).toEqual([0, 1])
  })

  it('pushes, removes and moves items', async () => {
    const form = useForm({ schema, initialValues: { items: [{ name: 'a' }] } })
    const items = form.array('items')
    items.push({ name: 'b' })
    await nextTick()
    expect(form.values.value.items).toEqual([{ name: 'a' }, { name: 'b' }])
    items.move(0, 1)
    await nextTick()
    expect(form.values.value.items).toEqual([{ name: 'b' }, { name: 'a' }])
    items.remove(0)
    await nextTick()
    expect(form.values.value.items).toEqual([{ name: 'a' }])
  })

  it('addresses nested array item fields by path', () => {
    const form = useForm({ schema, initialValues: { items: [{ name: 'a' }] } })
    expect(form.field('items.0.name').value.value).toBe('a')
  })

  it('memoizes array() per path — repeated calls return the same object', () => {
    const form = useForm({ schema, initialValues: { items: [] } })
    expect(form.array('items')).toBe(form.array('items'))
  })
})
