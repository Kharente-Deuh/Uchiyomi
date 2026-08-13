// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import PageLayout from './PageLayout.vue'

type Props = InstanceType<typeof PageLayout>['$props']

const DEFAULT_SLOTS = { default: () => 'page body' }

function wrap(
  props: Props,
  slots: Record<string, () => unknown> = DEFAULT_SLOTS,
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(PageLayout, props, slots)]) }
}

describe('organismPageLayout', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID' }))
    expect(wrapper.text()).toContain('PocketID')
  })

  it('renders the default slot', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID' }))
    expect(wrapper.text()).toContain('page body')
  })

  it('renders the subtitle when given one', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID', subtitle: 'OIDC provider' }))
    expect(wrapper.text()).toContain('OIDC provider')
  })

  it('renders no back link without back routes', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID' }))
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('renders no back link for an empty back route list', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID', backRoutes: [] }))
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('renders the back link with its name', async () => {
    const wrapper = await mountSuspended(wrap({
      title: 'PocketID',
      backRoutes: [{ to: '/settings/oidc', name: 'SSO' }],
    }))

    expect(wrapper.find('a').attributes('href')).toBe('/settings/oidc')
    expect(wrapper.text()).toContain('SSO')
  })

  it('renders every back route in order', async () => {
    const wrapper = await mountSuspended(wrap({
      title: 'PocketID',
      backRoutes: [
        { to: '/settings', name: 'Settings' },
        { to: '/settings/oidc', name: 'SSO' },
      ],
    }))

    expect(wrapper.findAll('a').map(a => a.attributes('href'))).toEqual(['/settings', '/settings/oidc'])
    expect(wrapper.findAll('a').map(a => a.text())).toEqual(['Settings', 'SSO'])
  })

  it('shows a progress bar while loading', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID', loading: true }))
    expect(wrapper.find('.v-progress-linear').exists()).toBe(true)
  })

  it('replaces the whole page with a spinner while the global loader is on', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID', loading: true, globalLoader: true }))

    expect(wrapper.find('.v-progress-circular').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('page body')
  })

  it('renders the page once the global loader is off', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'PocketID', globalLoader: true }))
    expect(wrapper.text()).toContain('page body')
  })

  it('renders the prepend-title slot next to the title', async () => {
    const wrapper = await mountSuspended(wrap(
      { title: 'PocketID' },
      { 'default': () => 'page body', 'prepend-title': () => 'logo' },
    ))

    expect(wrapper.text()).toContain('logo')
    expect(wrapper.text()).toContain('PocketID')
  })

  it('renders the sub-header slot', async () => {
    const wrapper = await mountSuspended(wrap(
      { title: 'Asura Scans' },
      { 'default': () => 'page body', 'sub-header': () => 'filters' },
    ))

    expect(wrapper.text()).toContain('filters')
  })
})
