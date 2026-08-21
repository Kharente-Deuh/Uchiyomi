// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { FeedItem } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import ComicCard from './index.vue'

function item(overrides: Partial<FeedItem> = {}): FeedItem {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    cover: '/cover/solo-leveling',
    source: 'asurascans',
    status: 'ongoing',
    type: 'manhwa',
    hasProgress: true,
    latestChapters: [{
      id: 'ch-1',
      number: 12,
      download: 100,
      publishedAt: new Date('2026-01-01'),
    }],
    ...overrides,
  }
}

async function mount(feedItem: FeedItem = item()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(ComicCard, { item: feedItem })]),
  })
}

describe('feedComicCard', () => {
  it('renders the title', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Solo Leveling')
  })

  it('links to the library comic page', async () => {
    const wrapper = await mount()
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1?from=feed')
  })

  it('renders each latest chapter', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Chapter 12')
  })
})
