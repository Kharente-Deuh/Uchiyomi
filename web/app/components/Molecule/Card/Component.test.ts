// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Component from './Component.vue'

type Props = InstanceType<typeof Component>['$props']

function wrap(props: Props, slot = 'body'): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Component, props, () => slot)]) }
}

const base = { title: 'Claims', icon: 'fa6-solid:key' }

describe('moleculeCardComponent', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('Claims')
  })

  it('renders the default slot', async () => {
    const wrapper = await mountSuspended(wrap(base, 'a form'))
    expect(wrapper.text()).toContain('a form')
  })

  it('is neither disabled nor loading by default', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.find('.v-card--disabled').exists()).toBe(false)
    expect(wrapper.find('.v-card--loading').exists()).toBe(false)
  })

  it('marks the card disabled', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, disabled: true }))
    expect(wrapper.find('.v-card--disabled').exists()).toBe(true)
  })

  it('marks the card loading', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, loading: true }))
    expect(wrapper.find('.v-card--loading').exists()).toBe(true)
  })
})
