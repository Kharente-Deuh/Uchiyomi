// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { Ref } from 'vue'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Provider from './Provider.vue'

function i18nStub(): { locale: Ref<string> } {
  return { locale: ref('en') }
}

mockNuxtImport('useI18n', () => i18nStub)

const base = {
  id: 'p1',
  displayName: 'PocketID',
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  userCount: 3,
}

function wrap(props: typeof base): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Provider, props)]) }
}

describe('oidcCardProvider', () => {
  it('renders the display name', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('PocketID')
  })

  it('renders the user count', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('3')
  })

  it('renders a zero user count rather than hiding it', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, userCount: 0 }))
    expect(wrapper.text()).toContain('0')
  })

  it('formats the creation date in the active locale', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain(base.createdAt.toLocaleDateString('en'))
  })

  it('links to the provider detail page', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.find('a').attributes('href')).toBe('/settings/oidc/p1')
  })
})
