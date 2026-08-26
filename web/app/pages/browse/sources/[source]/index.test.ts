// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceSearchItem } from '~/features/sources/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import SourceBrowsePage from './index.vue'

const { series, isLoading, hasNextPage, smAndDown, page, addComicInLibrary, addComicInLibraryLoading, infosLoading, resetFilters, params, navigateTo } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    series: ref<SourceSearchItem[]>([]),
    isLoading: ref(false),
    hasNextPage: ref(false),
    smAndDown: ref(false),
    page: ref(1),
    addComicInLibrary: vi.fn(),
    addComicInLibraryLoading: ref<Record<string, boolean>>({}),
    infosLoading: ref<Record<string, boolean>>({}),
    resetFilters: vi.fn(),
    params: { source: 'asurascans' },
    navigateTo: vi.fn(),
  }
})

vi.mock('~/features/sources/composables/sources-search.composable', () => ({
  useSourceSearch: () => ({
    isLoading,
    series,
    page,
    hasNextPage,
    addComicInLibrary,
    addComicInLibraryLoading,
    infosLoading,
    resetFilters,
  }),
}))

vi.mock('@vueuse/core', async importOriginal => ({
  ...(await importOriginal<typeof import('@vueuse/core')>()),
  useIntersectionObserver: vi.fn(),
}))

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()

  return {
    ...actual,
    onBeforeRouteLeave: vi.fn(),
  }
})

function routeStub(): { params: { source: string }, name: string } {
  return { params, name: 'browse-sources-source' }
}

function displayStub(): { smAndDown: typeof smAndDown } {
  return { smAndDown }
}

mockNuxtImport('useRoute', () => routeStub)
mockNuxtImport('useDisplay', () => displayStub)
mockNuxtImport('navigateTo', () => navigateTo)

const HeaderStub = defineComponent({
  name: 'SourcesHeader',
  template: '<div data-test="sources-header" />',
})

const CardStub = defineComponent({
  name: 'SourcesComicCard',
  props: {
    comic: { type: Object, required: true },
    sourceId: { type: String, required: true },
    statusLoading: { type: Boolean, default: false },
    chapterCountLoading: { type: Boolean, default: false },
  },
  emits: ['toggle'],
  template: '<button data-test="comic-card" :data-status-loading="statusLoading" :data-chapter-count-loading="chapterCountLoading" @click="$emit(\'toggle\')">{{ comic.title }}</button>',
})

const DeleteStub = defineComponent({
  name: 'SourcesModalDelete',
  template: '<div data-test="delete-modal" />',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(SourceBrowsePage)]) },
    {
      global: {
        stubs: {
          SourcesHeader: HeaderStub,
          SourcesComicCard: CardStub,
          SourcesModalDelete: DeleteStub,
          MoleculePaginationFooter: true,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

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
    chapterCount: 1,
    rating: 0,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  }
}

beforeEach(() => {
  series.value = []
  isLoading.value = false
  hasNextPage.value = false
  smAndDown.value = false
  page.value = 1
  infosLoading.value = {}
  params.source = ASURA_SOURCE_NAME
  addComicInLibrary.mockReset()
  resetFilters.mockReset()
  navigateTo.mockReset()
})

describe('browse Source Page', () => {
  it('shows an empty state when there are no series', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No results')
  })

  it('renders a card per series', async () => {
    series.value = [item()]
    const wrapper = await mount()
    expect(wrapper.findAll('[data-test="comic-card"]').map(n => n.text())).toEqual(['Solo Leveling'])
  })

  it('forwards infos loading onto each card', async () => {
    series.value = [item()]
    infosLoading.value = { 'solo-leveling': true }
    const wrapper = await mount()
    const card = wrapper.find('[data-test="comic-card"]')
    expect(card.attributes('data-status-loading')).toBe('true')
    expect(card.attributes('data-chapter-count-loading')).toBe('true')
  })

  it('adds a comic that is not in the library', async () => {
    const comic = item()
    series.value = [comic]
    const wrapper = await mount()

    await wrapper.find('[data-test="comic-card"]').trigger('click')

    expect(addComicInLibrary).toHaveBeenCalledWith(comic)
  })

  it('redirects when the source is unknown', async () => {
    params.source = 'unknown'

    await expect(mount()).rejects.toBeTruthy()
    expect(navigateTo).toHaveBeenCalledWith('/browse/sources', { replace: true })
  })
})
