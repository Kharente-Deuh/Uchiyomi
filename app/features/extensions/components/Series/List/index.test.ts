// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { SourceSearchItemDto, SourceSearchQueryType } from '#shared/dto/catalogue/source-search.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import PaginationFooter from '~/components/Molecule/PaginationFooter.vue'
import FiltersModal from '../Filters/Modal.vue'
import DesktopHeader from './Desktop/Header.vue'
import SeriesList from './index.vue'
import SeriesListItem from './Item.vue'
import MobileHeader from './Mobile/Header.vue'

// --- controlled composable state ---
// Defined at module scope (not vi.hoisted) so vue's ref is available; the
// mockNuxtImport factory only reads `state` lazily when the mock is invoked.
const state = {
  searchType: ref<SourceSearchQueryType>('popular'),
  selectedSourceId: ref<string | undefined>('src-1'),
  sourcesItems: ref<{ value: string, title: string, supportsLatest: boolean }[]>([]),
  series: ref<SourceSearchItemDto[]>([]),
  page: ref(1),
  hasNextPage: ref(false),
  searchFilter: ref<string>(),
  searchLoading: ref(false),
  definitions: ref<SourceFilterDto[]>([]),
  filtersDraft: ref<Record<string, FilterDraftValue>>({}),
  filtersLoading: ref(false),
}

function useSearchSeriesSourcesComposableMock(): unknown {
  return {
    searchType: state.searchType,
    selectedSourceId: state.selectedSourceId,
    sourcesItems: state.sourcesItems,
    series: computed(() => state.series.value),
    chapterStatusOf: vi.fn(),
    chapterSummaryOf: vi.fn(),
    page: state.page,
    hasNextPage: state.hasNextPage,
    searchFilter: state.searchFilter,
    searchLoading: state.searchLoading,
    definitions: state.definitions,
    filtersDraft: state.filtersDraft,
    filtersActiveCount: computed(() => 0),
    filtersLoading: state.filtersLoading,
    applyFilters: vi.fn(),
    resetFilters: vi.fn(),
  }
}

mockNuxtImport('useSearchSeriesSourcesComposable', () => useSearchSeriesSourcesComposableMock)

// --- useDisplay: toggle mobile per test ---
const mobileRef = ref(false)

function useDisplayStub(): { mobile: typeof mobileRef } {
  return { mobile: mobileRef }
}

mockNuxtImport('useDisplay', () => useDisplayStub)

// --- vueuse: useIntersectionObserver is a no-op under jsdom ---
vi.mock('@vueuse/core', async (importOriginal) => {
  const orig = await importOriginal<typeof import('@vueuse/core')>()

  return {
    ...orig,
    useIntersectionObserver: () => ({ stop: () => {} }),
  }
})

function makeManga(id: string): SourceSearchItemDto {
  return {
    id,
    title: `Manga ${id}`,
    thumbnailUrl: null,
    inLibrary: false,
    sourceUrl: null,
  }
}

function wrap(): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [h(SeriesList)]),
  }
}

describe('extensionsSeriesList', () => {
  beforeEach(() => {
    mobileRef.value = false
    state.searchType.value = 'popular'
    state.selectedSourceId.value = 'src-1'
    state.sourcesItems.value = []
    state.series.value = []
    state.page.value = 1
    state.hasNextPage.value = false
    state.searchFilter.value = undefined
    state.searchLoading.value = false
    state.definitions.value = []
    state.filtersDraft.value = {}
    state.filtersLoading.value = false
  })

  // --- searchTypeItems computed ---

  it('includes the latest option when the selected source supportsLatest', async () => {
    state.sourcesItems.value = [{ value: 'src-1', title: 'English', supportsLatest: true }]
    const wrapper = await mountSuspended(wrap())
    const items = wrapper.findComponent(DesktopHeader).props('searchTypeItems') as { title: string, value: string }[]
    const values = items.map(i => i.value)
    expect(values).toContain('popular')
    expect(values).toContain('latest')
    expect(values).toContain('search')
    const titles = items.map(i => i.title)
    expect(titles).toContain('Latest')
    expect(titles).toContain('Popular')
    expect(titles).toContain('Search')
  })

  it('omits the latest option when the selected source does not support latest', async () => {
    state.sourcesItems.value = [{ value: 'src-1', title: 'English', supportsLatest: false }]
    const wrapper = await mountSuspended(wrap())
    const items = wrapper.findComponent(DesktopHeader).props('searchTypeItems') as { title: string, value: string }[]
    const values = items.map(i => i.value)
    expect(values).toContain('popular')
    expect(values).not.toContain('latest')
    expect(values).toContain('search')
  })

  // --- desktop vs mobile layout ---

  it('renders the desktop header and pagination footer on desktop', async () => {
    mobileRef.value = false
    state.hasNextPage.value = true
    state.series.value = [makeManga('m1')]
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(DesktopHeader).exists()).toBe(true)
    expect(wrapper.findComponent(MobileHeader).exists()).toBe(false)
    expect(wrapper.findComponent(PaginationFooter).exists()).toBe(true)
    expect(wrapper.find('.series-list-grid').classes()).not.toContain('px-4')
  })

  it('renders the mobile header and no pagination footer on mobile', async () => {
    mobileRef.value = true
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(MobileHeader).exists()).toBe(true)
    expect(wrapper.findComponent(DesktopHeader).exists()).toBe(false)
    expect(wrapper.findComponent(PaginationFooter).exists()).toBe(false)
  })

  // --- series grid ---

  it('renders one list item per series entry', async () => {
    state.series.value = [makeManga('m1'), makeManga('m2')]
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findAllComponents(SeriesListItem)).toHaveLength(2)
  })

  it('renders the filters modal', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(FiltersModal).exists()).toBe(true)
  })
})
