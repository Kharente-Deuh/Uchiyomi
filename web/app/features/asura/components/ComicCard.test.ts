// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraSearchItem } from '../composables/asura.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import ComicCard from './ComicCard.vue'

function item(internalId?: string): AsuraSearchItem {
  return {
    slug: 'solo-leveling',
    title: 'Solo Leveling',
    cover: '/api/sources/cover/solo-leveling?source=asurascans',
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: '',
    artist: '',
    description: '',
    altTitles: [],
    genres: [],
    latestChapters: [],
    chapterCount: 12,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...(internalId && { internalId }),
  }
}

async function mount(comic: AsuraSearchItem, isLoading = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(ComicCard, { comic, loading: isLoading })]),
  })
}

describe('asuraComicCard', () => {
  it('renders the title', async () => {
    const wrapper = await mount(item())
    expect(wrapper.text()).toContain('Solo Leveling')
  })

  it('links to the series page', async () => {
    const wrapper = await mount(item())
    expect(wrapper.find('a').attributes('href')).toBe('/browse/sources/asura/solo-leveling')
  })

  it('shows the add button when the comic is not in the library', async () => {
    const wrapper = await mount(item())
    expect(wrapper.find('.add-library-btn').exists()).toBe(true)
    expect(wrapper.find('.remove-library-btn').exists()).toBe(false)
  })

  it('shows the remove control when the comic is in the library', async () => {
    const wrapper = await mount(item('c1'))
    expect(wrapper.find('.remove-library-btn').exists()).toBe(true)
    expect(wrapper.find('.add-library-btn').exists()).toBe(false)
  })
})
