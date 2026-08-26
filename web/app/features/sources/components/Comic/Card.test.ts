// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceSearchItem } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Card from './Card.vue'

function item(overrides: Partial<SourceSearchItem> = {}): SourceSearchItem {
  return {
    slug: 'solo-leveling',
    title: 'Solo Leveling',
    cover: '/cover',
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: '',
    artist: '',
    description: '',
    altTitles: [],
    genres: [],
    chapterCount: 12,
    rating: 0,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  }
}

async function mount(comic: SourceSearchItem = item(), isLoading = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Card, {
      sourceId: ASURA_SOURCE_NAME,
      comic,
      loading: isLoading,
    })]),
  })
}

describe('sourcesComicCard', () => {
  it('renders the title and chapter count', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Solo Leveling')
    expect(wrapper.text()).toContain('12 Chapters')
  })

  it('links to the source comic page', async () => {
    const wrapper = await mount()
    expect(wrapper.find('a').attributes('href')).toBe('/browse/sources/asurascans/solo-leveling')
  })

  it('emits toggle from the add button', async () => {
    const wrapper = await mount()

    await wrapper.find('.add-library-btn').trigger('click')

    expect(wrapper.findComponent(Card).emitted('toggle')).toHaveLength(1)
  })

  it('emits toggle from the delete button when the comic is in the library', async () => {
    const wrapper = await mount(item({ internalId: 'c1' }))

    await wrapper.find('.remove-library-btn').trigger('click')

    expect(wrapper.findComponent(Card).emitted('toggle')).toHaveLength(1)
  })
})
