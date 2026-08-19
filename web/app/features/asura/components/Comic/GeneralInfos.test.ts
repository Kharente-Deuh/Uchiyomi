// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraComicInfos } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import GeneralInfos from './GeneralInfos.vue'

function comic(overrides: Partial<AsuraComicInfos> = {}): AsuraComicInfos {
  return {
    title: 'Solo Leveling',
    cover: '/cover',
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    description: '<p>A hunter.</p>\nSecond line',
    slug: 'solo-leveling',
    altTitles: ['나 혼자만 레벨업', 'Only I Level Up'],
    genres: ['action'],
    chapterCount: 1,
    rating: 9,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  }
}

async function mount(value: AsuraComicInfos = comic()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(GeneralInfos, { comic: value })]),
  })
}

describe('asuraComicGeneralInfos', () => {
  it('strips HTML and newlines from the description', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('A hunter. Second line')
    expect(wrapper.text()).not.toContain('<p>')
  })

  it('joins alternate titles', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('나 혼자만 레벨업, Only I Level Up')
  })

  it('expands truncated text when show more is clicked', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Show more')
    expect(wrapper.find('.text-truncate').exists()).toBe(true)

    await wrapper.findAll('button').at(-1)!.trigger('click')

    expect(wrapper.text()).toContain('Show less')
    expect(wrapper.find('.text-truncate').exists()).toBe(false)
  })
})
