// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceDto } from '#shared/dto/extensions'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VSwitch } from 'vuetify/components'
import Item from './Item.vue'

const baseSource: SourceDto = {
  id: 'src-1',
  name: 'Example Source',
  lang: 'en',
  isNsfw: false,
  isConfigurable: false,
  isEnabled: false,
  supportsLatest: true,
}

function wrap(
  source: SourceDto,
  canManage = false,
  loading = false,
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [
      h(Item, {
        source,
        canManage,
        loading,
        onToggle: () => {},
        onSettings: () => {},
      }),
    ]),
  }
}

describe('extensionsSourceListDesktopItem', () => {
  it('shows the language display text', async () => {
    const wrapper = await mountSuspended(wrap(baseSource))
    // 'en' maps to English via endomyLanguage
    expect(wrapper.text()).toContain('English')
  })

  it('applies primary text colour when the source is enabled', async () => {
    const wrapper = await mountSuspended(wrap({ ...baseSource, isEnabled: true }))
    expect(wrapper.find('.text-primary').exists()).toBe(true)
  })

  it('shows the switch when canManage is true', async () => {
    const wrapper = await mountSuspended(wrap(baseSource, true))
    expect(wrapper.findComponent(VSwitch).exists()).toBe(true)
  })

  it('does not show the switch when canManage is false', async () => {
    const wrapper = await mountSuspended(wrap(baseSource, false))
    expect(wrapper.findComponent(VSwitch).exists()).toBe(false)
  })

  it('emits toggle when the card is clicked', async () => {
    const onToggle = vi.fn()
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(Item, { source: baseSource, onToggle }),
      ]),
    })
    await wrapper.find('.source-card').trigger('click')
    expect(onToggle).toHaveBeenCalled()
  })

  it('disables the switch when loading is true', async () => {
    const wrapper = await mountSuspended(wrap(baseSource, true, true))
    expect(wrapper.findComponent(VSwitch).props('disabled')).toBe(true)
  })
})
