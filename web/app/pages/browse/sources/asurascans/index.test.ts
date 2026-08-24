// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraScansSearchItem } from '~/features/asurascans/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import AsuraIndex from './index.vue'

const { series, isLoading, maxPage, smAndDown, page, addComicInLibraryLoading } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    series: ref<AsuraScansSearchItem[]>([]),
    isLoading: ref(false),
    maxPage: ref(1),
    smAndDown: ref(false),
    page: ref(1),
    addComicInLibraryLoading: ref<Record<string, boolean>>({}),
  }
})

vi.mock('~/features/asurascans/composables/asurascans-search.composable', () => ({
  useAsuraScansSearch: () => ({
    isLoading,
    series,
    page,
    maxPage,
    addComicInLibrary: vi.fn(),
    addComicInLibraryLoading,
  }),
}))

vi.mock('@vueuse/core', async importOriginal => ({
  ...(await importOriginal<typeof import('@vueuse/core')>()),
  useIntersectionObserver: vi.fn(),
}))

const { mockNuxtImport } = await import('@nuxt/test-utils/runtime')

function displayStub(): { smAndDown: typeof smAndDown } {
  return { smAndDown }
}

mockNuxtImport('useDisplay', () => displayStub)

const HeaderStub = defineComponent({
  name: 'AsuraScansHeader',
  template: '<div data-test="asura-header" />',
})

const CardStub = defineComponent({
  name: 'AsuraScansComicCard',
  props: { comic: { type: Object, required: true } },
  template: '<div data-test="comic-card">{{ comic.title }}</div>',
})

const DeleteStub = defineComponent({
  name: 'AsuraScansModalDelete',
  template: '<div data-test="delete-modal" />',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(AsuraIndex)]) },
    {
      global: {
        stubs: {
          AsuraScansHeader: HeaderStub,
          AsuraScansComicCard: CardStub,
          AsuraScansModalDelete: DeleteStub,
          MoleculePaginationFooter: true,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

function item(): AsuraScansSearchItem {
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
    latestChapters: [],
    chapterCount: 1,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
  }
}

beforeEach(() => {
  series.value = []
  isLoading.value = false
  maxPage.value = 1
  smAndDown.value = false
})

describe('asuraScansBrowsePage', () => {
  it('shows an empty state when there are no series', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No results')
  })

  it('renders a card per series', async () => {
    series.value = [item()]
    const wrapper = await mount()
    expect(wrapper.findAll('[data-test="comic-card"]').map(n => n.text())).toEqual(['Solo Leveling'])
  })
})
