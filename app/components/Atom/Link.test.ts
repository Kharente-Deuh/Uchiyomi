// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import Link from './Link.vue'

function wrap(
  props: Record<string, unknown> = {},
  slot = 'LINK_SLOT',
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(Link, props, { default: () => slot }) }
}

describe('atomLink', () => {
  it('renders an anchor to the target with the slot content', async () => {
    const wrapper = await mountSuspended(wrap({ to: '/library' }))
    const anchor = wrapper.find('a')
    expect(anchor.exists()).toBe(true)
    expect(anchor.attributes('href')).toBe('/library')
    expect(wrapper.text()).toContain('LINK_SLOT')
  })

  it('renders only the slot (no anchor) when no target is given', async () => {
    const wrapper = await mountSuspended(wrap({}))
    expect(wrapper.find('a').exists()).toBe(false)
    expect(wrapper.text()).toContain('LINK_SLOT')
  })

  it('opens in a new tab when newTab is set', async () => {
    const wrapper = await mountSuspended(wrap({ to: '/x', newTab: true }))
    expect(wrapper.find('a').attributes('target')).toBe('_blank')
  })

  it('strips the text decoration by default and keeps it when enabled', async () => {
    const off = await mountSuspended(wrap({ to: '/x' }))
    expect(off.find('a').classes()).toContain('text-decoration-none')

    const on = await mountSuspended(wrap({ to: '/x', enableTextDecoration: true }))
    expect(on.find('a').classes()).not.toContain('text-decoration-none')
  })
})
