// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceDto } from '#shared/dto/extensions'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import DesktopList from './Desktop/index.vue'
import SourceList from './index.vue'
import MobileList from './Mobile/index.vue'

const mobileRef = ref(false)

function useDisplayStub(): { mobile: typeof mobileRef } {
  return { mobile: mobileRef }
}

mockNuxtImport('useDisplay', () => useDisplayStub)

function makeSource(id: string, isEnabled = false): SourceDto {
  return {
    id,
    name: `Source ${id}`,
    lang: 'en',
    isNsfw: false,
    isConfigurable: false,
    isEnabled,
    supportsLatest: true,
  }
}

const sources: SourceDto[] = [makeSource('a1'), makeSource('a2', true)]

function wrap(): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [
      h(SourceList, {
        sources,
        canManageExtensions: false,
        sourceToggleLoading: new Set<string>(),
        hasSettings: false,
        pkgName: 'com.example.pkg',
        onToggle: () => {},
      }),
    ]),
  }
}

describe('extensionsSourceList', () => {
  it('renders the desktop list on desktop', async () => {
    mobileRef.value = false
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(DesktopList).exists()).toBe(true)
    expect(wrapper.findComponent(MobileList).exists()).toBe(false)
  })

  it('renders the mobile list on mobile', async () => {
    mobileRef.value = true
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(MobileList).exists()).toBe(true)
    expect(wrapper.findComponent(DesktopList).exists()).toBe(false)
  })

  it('computes enabledSourcesCount correctly and passes it down', async () => {
    mobileRef.value = false
    const wrapper = await mountSuspended(wrap())
    const desktop = wrapper.findComponent(DesktopList)
    // One source is enabled (a2)
    expect(desktop.props('enabledSourcesCount')).toBe(1)
  })

  it('passes totalSourcesCount equal to sources.length', async () => {
    mobileRef.value = false
    const wrapper = await mountSuspended(wrap())
    const desktop = wrapper.findComponent(DesktopList)
    expect(desktop.props('totalSourcesCount')).toBe(2)
  })
})
