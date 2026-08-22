// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { FeedChapter } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VProgressCircular } from 'vuetify/components'
import Chapter from './Chapter.vue'

function chapter(overrides: Partial<FeedChapter> = {}): FeedChapter {
  return {
    id: 'ch-1',
    number: 12,
    download: 100,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    hasProgress: false,
    ...overrides,
  }
}

async function mount(value: FeedChapter): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Chapter, { comicId: 'comic-1', chapter: value })]),
  })
}

describe('feedComicCardChapter', () => {
  it('renders the chapter number', async () => {
    const wrapper = await mount(chapter())
    expect(wrapper.text()).toContain('Chapter 12')
  })

  it('marks a completed download as hoverable', async () => {
    const wrapper = await mount(chapter({ download: 100 }))
    expect(wrapper.find('.feed-chapter').exists()).toBe(true)
  })

  it('shows an error icon when download failed', async () => {
    const wrapper = await mount(chapter({ download: -1 }))
    expect(wrapper.find('.feed-chapter').exists()).toBe(false)
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('shows a progress ring while downloading', async () => {
    const wrapper = await mount(chapter({ download: 40 }))
    const progress = wrapper.findComponent(VProgressCircular)

    expect(progress.exists()).toBe(true)
    expect(progress.props('indeterminate')).toBe(false)
  })
})
