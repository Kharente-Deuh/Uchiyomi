// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraComicInfos } from '../../types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import StatusInfos from './StatusInfos.vue'

function asuraSearchStub(): { addComicInLibrary: () => Promise<undefined> } {
  return { addComicInLibrary: vi.fn() }
}

function asuraChaptersStub(): { fetchChapters: () => Promise<void> } {
  return { fetchChapters: vi.fn() }
}

mockNuxtImport('useAsuraSearch', () => asuraSearchStub)
mockNuxtImport('useAsuraChapters', () => asuraChaptersStub)

const DeleteStub = defineComponent({
  name: 'AsuraModalDelete',
  props: {
    modelValue: { type: Boolean, default: false },
    comic: { type: Object, required: true },
  },
  template: '<div data-test="delete-modal" />',
})

const LinkStub = defineComponent({
  name: 'AsuraComicLinkOriginalSite',
  props: { to: { type: String, required: true } },
  template: '<a data-test="origin" :href="to" />',
})

function infos(overrides: Partial<AsuraComicInfos> = {}): AsuraComicInfos {
  return {
    title: 'Solo Leveling',
    cover: '/cover',
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    description: 'A hunter',
    slug: 'solo-leveling',
    altTitles: [],
    genres: ['action', 'fantasy'],
    chapterCount: 1,
    rating: 9,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  }
}

async function mount(comic: AsuraComicInfos): Promise<{ wrapper: VueWrapper, comic: AsuraComicInfos }> {
  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(StatusInfos, {
        comicOriginUrl: 'https://asurascans.com/series/solo-leveling',
        modelValue: comic,
      })]),
    },
    {
      global: {
        stubs: {
          AsuraModalDelete: DeleteStub,
          AsuraComicLinkOriginalSite: LinkStub,
        },
      },
    },
  )

  return { wrapper, comic }
}

describe('asuraComicStatusInfos', () => {
  it('renders status, type, author, artist and genres', async () => {
    const { wrapper } = await mount(infos())

    expect(wrapper.text()).toContain('Ongoing')
    expect(wrapper.text()).toContain('Manhwa')
    expect(wrapper.text()).toContain('Chugong')
    expect(wrapper.text()).toContain('Jang')
    expect(wrapper.text()).toContain('action')
    expect(wrapper.text()).toContain('fantasy')
    expect(wrapper.find('[data-test="origin"]').attributes('href')).toBe('https://asurascans.com/series/solo-leveling')
  })

  it('shows the add-to-library action when the comic is not in the library', async () => {
    const { wrapper } = await mount(infos())

    expect(wrapper.text()).toContain('Add to library')
    expect(wrapper.findComponent(VBtn).props('color')).toBe('primary')
  })

  it('shows the remove action when the comic is already in the library', async () => {
    const { wrapper } = await mount(infos({ internalId: 'c1' }))

    expect(wrapper.text()).toContain('Remove from library')
    expect(wrapper.find('[data-test="delete-modal"]').exists()).toBe(true)
  })
})
