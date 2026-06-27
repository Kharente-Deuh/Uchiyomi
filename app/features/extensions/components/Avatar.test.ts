// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VImg } from 'vuetify/components'
import Avatar from './Avatar.vue'

function wrap(props: { url: string, size?: 'small' | 'default' | 'large' }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Avatar, props)]) }
}

describe('extensionsAvatar', () => {
  it('passes the url to VImg src', async () => {
    const wrapper = await mountSuspended(wrap({ url: 'https://example.com/icon.png' }))
    expect(wrapper.findComponent(VImg).props('src')).toBe('https://example.com/icon.png')
  })

  it('uses 48px width for size small', async () => {
    const wrapper = await mountSuspended(wrap({ url: '', size: 'small' }))
    expect(wrapper.findComponent(VImg).props('width')).toBe('48px')
  })

  it('uses 80px width for size large', async () => {
    const wrapper = await mountSuspended(wrap({ url: '', size: 'large' }))
    expect(wrapper.findComponent(VImg).props('width')).toBe('80px')
  })

  it('uses 52px width for the default size', async () => {
    const wrapper = await mountSuspended(wrap({ url: '' }))
    expect(wrapper.findComponent(VImg).props('width')).toBe('52px')
  })

  it('applies a 16px border-radius for size large', async () => {
    const wrapper = await mountSuspended(wrap({ url: '', size: 'large' }))
    const div = wrapper.find('.extension-avatar')
    expect(div.attributes('style')).toContain('border-radius: 16px')
  })

  it('applies an 8px border-radius for non-large sizes', async () => {
    const wrapper = await mountSuspended(wrap({ url: '', size: 'small' }))
    const div = wrapper.find('.extension-avatar')
    expect(div.attributes('style')).toContain('border-radius: 8px')
  })
})
