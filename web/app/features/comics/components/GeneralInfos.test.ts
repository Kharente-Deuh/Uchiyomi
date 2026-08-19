// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Comic } from '../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import GeneralInfos from './GeneralInfos.vue'

function comic(overrides: Partial<Comic> = {}): Comic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: ASURA_SOURCE_NAME,
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    description: '<p>A hunter.</p>\nSecond line',
    cover: '/cover',
    genres: ['action'],
    altTitles: ['나 혼자만 레벨업', 'Only I Level Up'],
    chapterCount: 1,
    ...overrides,
  }
}

async function mount(value: Comic = comic()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(GeneralInfos, { comic: value })]),
  })
}

describe('comicsGeneralInfos', () => {
  it('strips HTML and newlines from the description', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('A hunter. Second line')
    expect(wrapper.text()).not.toContain('<p>')
  })

  it('joins alternate titles', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('나 혼자만 레벨업, Only I Level Up')
  })

  it('hides genres until show more is clicked', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Show more')
    expect(wrapper.text()).not.toContain('action')
    expect(wrapper.find('.text-truncate').exists()).toBe(true)

    await wrapper.findAll('button').at(-1)!.trigger('click')

    expect(wrapper.text()).toContain('Show less')
    expect(wrapper.text()).toContain('action')
    expect(wrapper.find('.text-truncate').exists()).toBe(false)
  })
})
