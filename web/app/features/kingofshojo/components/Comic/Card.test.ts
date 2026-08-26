// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { KingOfShojoSearchItem } from '../../types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Card from './Card.vue'

const mocks = vi.hoisted(() => ({
  getInfosBySlug: vi.fn(),
}))

function createKingOfShojoApiMock(): { getInfosBySlug: typeof mocks.getInfosBySlug } {
  return { getInfosBySlug: mocks.getInfosBySlug }
}

mockNuxtImport('createKingOfShojoApi', () => createKingOfShojoApiMock)

const SourcesCardStub = defineComponent({
  name: 'SourcesCardComic',
  props: {
    status: { type: String, default: undefined },
    chapterCount: { type: Number, default: 0 },
    chapterCountLoading: { type: Boolean, default: false },
    statusLoading: { type: Boolean, default: false },
    internalId: { type: String, default: undefined },
    to: { type: String, required: true },
    cover: { type: String, required: true },
    title: { type: String, required: true },
    loading: { type: Boolean, default: false },
  },
  template: '<div data-test="sources-card" />',
})

function comic(overrides: Partial<KingOfShojoSearchItem> = {}): KingOfShojoSearchItem {
  return {
    title: 'Webtoon Character Na Kang Lim',
    cover: '/cover',
    publicUrl: '/manga/webtoon-character-na-kang-lim/',
    sourceUrl: '',
    status: '' as KingOfShojoSearchItem['status'],
    type: '' as KingOfShojoSearchItem['type'],
    author: '',
    artist: '',
    description: '',
    slug: 'webtoon-character-na-kang-lim',
    altTitles: [],
    genres: [],
    chapterCount: 0,
    rating: 0,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  } as KingOfShojoSearchItem
}

async function mount(item: KingOfShojoSearchItem = comic()): Promise<VueWrapper> {
  return mountSuspended(
    {
      render: () => h(VApp, () => [h(Card, { comic: item })]),
    },
    {
      global: {
        stubs: {
          SourcesCardComic: SourcesCardStub,
        },
      },
    },
  )
}

beforeEach(() => {
  mocks.getInfosBySlug.mockReset()
  mocks.getInfosBySlug.mockReturnValue(new Promise(() => {}))
})

describe('kingOfShojoComicCard', () => {
  it('does not invent an ongoing status before series infos load', async () => {
    const wrapper = await mount()

    expect(wrapper.findComponent(SourcesCardStub).props('status')).toBeUndefined()
  })

  it('uses the status returned by getInfosBySlug', async () => {
    mocks.getInfosBySlug.mockResolvedValue({
      success: true,
      data: { status: 'completed', chapterCount: 42 },
    })

    const wrapper = await mount()
    await vi.waitFor(() => {
      expect(wrapper.findComponent(SourcesCardStub).props('status')).toBe('completed')
    })
    expect(wrapper.findComponent(SourcesCardStub).props('chapterCount')).toBe(42)
  })
})
