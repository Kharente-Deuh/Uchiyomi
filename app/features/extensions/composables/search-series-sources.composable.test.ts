// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceSearchItemDto } from '#shared/dto/catalogue/source-search.dto'
import type { ExtensionDto, SourceDto } from '#shared/dto/extensions'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

const { mockApi } = vi.hoisted(() => ({
  mockApi: {
    searchSeriesBySource: vi.fn(),
    getMangaChapterSummary: vi.fn(),
    getSourceFilters: vi.fn(),
  },
}))

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => mockApi,
}))

// The composable `watch`es the `data` returned by useQuery, so it must be a real
// reactive ref for mutations to propagate — a plain `{ value }` stub is not
// tracked by Vue. vi.hoisted runs before imports (no `ref()` yet), so we hold a
// mutable container the mock factory fills with real refs created from `vue`.
const { queryHolder } = vi.hoisted(() => ({
  queryHolder: {} as {
    data: { value: unknown }
    isLoading: { value: boolean }
  },
}))

vi.mock('@pinia/colada', async () => {
  const { ref } = await import('vue')
  queryHolder.data = ref<unknown>(undefined)
  queryHolder.isLoading = ref(false)

  return {
    useQuery: () => queryHolder,
  }
})

// Mock @vueuse/core's useDebounce (used internally for the search filter) so it
// passes the ref through synchronously.
vi.mock('@vueuse/core', async (importOriginal) => {
  const orig = await importOriginal<typeof import('@vueuse/core')>()

  return {
    ...orig,
    useDebounce: (v: { value: unknown }) => v,
  }
})

const { successMock, errorMock } = vi.hoisted(() => ({
  successMock: vi.fn(),
  errorMock: vi.fn(),
}))

// Mutable mobile ref the tests flip to drive desktop/mobile accumulation.
const { mobileRef } = vi.hoisted(() => ({
  mobileRef: { value: false } as { value: boolean },
}))

function useToastMock(): { success: typeof successMock, error: typeof errorMock } {
  return { success: successMock, error: errorMock }
}

function useI18nMock(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function useDisplayMock(): { mobile: { value: boolean } } {
  return { mobile: mobileRef }
}

mockNuxtImport('useToast', () => useToastMock)
mockNuxtImport('useI18n', () => useI18nMock)
mockNuxtImport('useDisplay', () => useDisplayMock)

const { useSearchSeriesSourcesComposable } = await import('~/features/extensions/composables/search-series-sources.composable')
const { useSingleExtensionStore } = await import('~/features/extensions/store/single-extension.store')
const { useChapterSummariesStore } = await import('~/features/extensions/store/chapter-summaries.store')

// --- fixture data ---

const ext: ExtensionDto = {
  pkgName: 'eu.kanade.test',
  name: 'Test',
  lang: 'en',
  isNsfw: false,
  isInstalled: true,
  hasUpdate: false,
  versionName: '1.0',
}

function makeSource(overrides: Partial<SourceDto> = {}): SourceDto {
  return {
    id: 's1',
    name: 'TestSource',
    lang: 'en',
    isNsfw: false,
    isConfigurable: false,
    isEnabled: true,
    supportsLatest: true,
    ...overrides,
  }
}

function makeItem(id: string): SourceSearchItemDto {
  return {
    id,
    title: `Title ${id}`,
    thumbnailUrl: null,
    inLibrary: false,
    sourceUrl: null,
  }
}

describe('useSearchSeriesSourcesComposable', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockApi.getSourceFilters.mockResolvedValue({ success: true, data: [] })
    queryHolder.data.value = undefined
    queryHolder.isLoading.value = false
    mobileRef.value = false
    // Neutralise the real summaries worker: the `series` watcher calls
    // `summaries.sync`, whose worker would otherwise fetch through the (unmocked)
    // chapter-summary API and reject. Tests that assert on `sync` re-spy it.
    vi.spyOn(useChapterSummariesStore(), 'sync').mockImplementation(() => {})
  })

  // --- sourcesItems (#1) ---

  it('sourcesItems only includes enabled sources', () => {
    const store = useSingleExtensionStore()
    store.setSources([
      makeSource({ id: 's1', isEnabled: true }),
      makeSource({ id: 's2', isEnabled: false }),
    ])
    const { sourcesItems } = useSearchSeriesSourcesComposable()
    expect(sourcesItems.value).toHaveLength(1)
    expect(sourcesItems.value[0]!.value).toBe('s1')
  })

  it('sourcesItems maps lang "all" to the sources.all i18n key', () => {
    const store = useSingleExtensionStore()
    store.setSources([makeSource({ id: 's1', lang: 'all' })])
    const { sourcesItems } = useSearchSeriesSourcesComposable()
    expect(sourcesItems.value[0]!.title).toBe('sources.all')
  })

  it('sourcesItems maps a concrete lang via endomyLanguage', () => {
    const store = useSingleExtensionStore()
    store.setSources([makeSource({ id: 's1', lang: 'en' })])
    const { sourcesItems } = useSearchSeriesSourcesComposable()
    expect(sourcesItems.value[0]!.title).toBe('English')
  })

  it('sourcesItems carries supportsLatest per source', () => {
    const store = useSingleExtensionStore()
    store.setSources([makeSource({ id: 's1', supportsLatest: false })])
    const { sourcesItems } = useSearchSeriesSourcesComposable()
    expect(sourcesItems.value[0]!.supportsLatest).toBe(false)
  })

  it('sourcesItems is empty when there are no sources', () => {
    const { sourcesItems } = useSearchSeriesSourcesComposable()
    expect(sourcesItems.value).toEqual([])
  })

  // --- selectedSourceId (#2) ---

  it('selectedSourceId initialises to the first enabled source id', () => {
    const store = useSingleExtensionStore()
    store.setSources([
      makeSource({ id: 'disabled', isEnabled: false }),
      makeSource({ id: 'first-enabled', isEnabled: true }),
      makeSource({ id: 'second-enabled', isEnabled: true }),
    ])
    const { selectedSourceId } = useSearchSeriesSourcesComposable()
    expect(selectedSourceId.value).toBe('first-enabled')
  })

  // --- hasNextPage (#5) ---

  it('hasNextPage reflects data.hasNextPage', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { hasNextPage } = useSearchSeriesSourcesComposable()
    expect(hasNextPage.value).toBe(false)

    queryHolder.data.value = { hasNextPage: true, items: [] }
    await nextTick()
    expect(hasNextPage.value).toBe(true)
  })

  // --- accumulation (#3) ---

  it('replaces series on page 1 (desktop)', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { series, page } = useSearchSeriesSourcesComposable()

    page.value = 1
    queryHolder.data.value = { hasNextPage: true, items: [makeItem('a'), makeItem('b')] }
    await nextTick()
    expect(series.value.map(s => s.id)).toEqual(['a', 'b'])

    // A subsequent desktop page replaces (does not append).
    page.value = 2
    queryHolder.data.value = { hasNextPage: false, items: [makeItem('c')] }
    await nextTick()
    expect(series.value.map(s => s.id)).toEqual(['c'])
  })

  it('appends series on mobile when page > 1', async () => {
    mobileRef.value = true
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { series, page } = useSearchSeriesSourcesComposable()

    page.value = 1
    queryHolder.data.value = { hasNextPage: true, items: [makeItem('a'), makeItem('b')] }
    await nextTick()
    expect(series.value.map(s => s.id)).toEqual(['a', 'b'])

    page.value = 2
    queryHolder.data.value = { hasNextPage: true, items: [makeItem('c'), makeItem('d')] }
    await nextTick()
    expect(series.value.map(s => s.id)).toEqual(['a', 'b', 'c', 'd'])
  })

  it('resets series on mobile when page is 1', async () => {
    mobileRef.value = true
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { series, page } = useSearchSeriesSourcesComposable()

    page.value = 2
    queryHolder.data.value = { hasNextPage: true, items: [makeItem('a')] }
    await nextTick()
    page.value = 1
    queryHolder.data.value = { hasNextPage: true, items: [makeItem('b')] }
    await nextTick()
    expect(series.value.map(s => s.id)).toEqual(['b'])
  })

  // --- filters integration ---

  it('applyFilters forces the search type and resets page to 1', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { applyFilters, searchType, page } = useSearchSeriesSourcesComposable()

    page.value = 4
    searchType.value = 'popular'
    applyFilters()
    await nextTick()

    expect(searchType.value).toBe('search')
    expect(page.value).toBe(1)
  })

  it('exposes a filters active count starting at zero', () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const { filtersActiveCount } = useSearchSeriesSourcesComposable()
    expect(filtersActiveCount.value).toBe(0)
  })

  // --- summaries.sync (#4) ---

  it('series watcher syncs summaries with the current pkgName/sourceId', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([makeSource({ id: 's1' })])
    const summaries = useChapterSummariesStore()
    const syncSpy = vi.spyOn(summaries, 'sync').mockImplementation(() => {})

    useSearchSeriesSourcesComposable()

    queryHolder.data.value = { hasNextPage: false, items: [makeItem('a'), makeItem('b')] }
    await nextTick()

    expect(syncSpy).toHaveBeenCalledWith(['a', 'b'], { pkgName: ext.pkgName, sourceId: 's1' })
  })
})
