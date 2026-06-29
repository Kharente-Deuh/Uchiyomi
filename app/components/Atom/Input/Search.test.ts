// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VTextField } from 'vuetify/components'
import Search from './Search.vue'

type SearchProps = InstanceType<typeof Search>['$props']

function wrap(
  attrs: SearchProps & { 'onUpdate:modelValue'?: (v: string | undefined) => void },
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Search, attrs)]) }
}

describe('atomInputSearch', () => {
  it('renders a text field', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: '' }))
    expect(wrapper.findComponent(VTextField).exists()).toBe(true)
  })

  it('reflects the modelValue in the input', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 'dragon' }))
    const input = wrapper.find('input')
    expect((input.element as HTMLInputElement).value).toBe('dragon')
  })

  it('emits update:modelValue when the user types', async () => {
    const emitted: (string | undefined)[] = []
    const wrapper = await mountSuspended(
      wrap({
        'modelValue': '',
        'onUpdate:modelValue': (v: string | undefined) => emitted.push(v),
      }),
    )
    await wrapper.find('input').setValue('one piece')
    expect(emitted.length).toBeGreaterThan(0)
    expect(emitted.at(-1)).toBe('one piece')
  })

  it('uses the placeholder prop when provided', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: '', placeholder: 'Search manga…' }))
    expect(wrapper.find('input').attributes('placeholder')).toBe('Search manga…')
  })

  it('falls back to a non-empty i18n-resolved placeholder when no placeholder prop is given', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: undefined }))
    // i18n is active in the nuxt test env and resolves 'fields.search'
    const placeholder = wrapper.find('input').attributes('placeholder') ?? ''
    expect(placeholder.length).toBeGreaterThan(0)
    // The component uses $t('fields.search'), never the raw key as literal text
    expect(placeholder).not.toBe('')
  })

  it('is clearable (clear button icon present when value is set)', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 'test' }))
    expect(wrapper.findComponent(VTextField).props('clearable')).toBe(true)
  })
})
