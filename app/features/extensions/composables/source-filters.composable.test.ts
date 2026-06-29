// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref } from 'vue'

const { mockApi } = vi.hoisted(() => ({
  mockApi: { getSourceFilters: vi.fn() },
}))

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => mockApi,
}))

const { errorMock } = vi.hoisted(() => ({ errorMock: vi.fn() }))

function useToastMock(): { success: ReturnType<typeof vi.fn>, error: ReturnType<typeof vi.fn> } {
  return { success: vi.fn(), error: errorMock }
}

function useI18nMock(): { t: (k: string) => string } {
  return { t: (k: string) => k }
}

mockNuxtImport('useToast', () => useToastMock)
mockNuxtImport('useI18n', () => useI18nMock)

const { useSourceFilters, filterPath } = await import('~/features/extensions/composables/source-filters.composable')

const DEFS: SourceFilterDto[] = [
  { type: 'checkbox', position: 0, name: 'Completed', default: false },
  { type: 'group', position: 1, name: 'Genres', filters: [
    { type: 'tristate', position: 0, name: 'Comedy', default: 'IGNORE' },
  ] },
]

describe('useSourceFilters', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockApi.getSourceFilters.mockResolvedValue({ success: true, data: DEFS })
  })

  it('does not fetch when enabled is false, even if a sourceId is set', async () => {
    const sourceId = ref<string | undefined>('s1')
    const enabled = ref(false)
    const { definitions } = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled })
    await nextTick()
    await nextTick()
    expect(mockApi.getSourceFilters).not.toHaveBeenCalled()
    expect(definitions.value).toEqual([])
  })

  it('fetches once enabled becomes true with a source set, then memoizes per sourceId', async () => {
    const sourceId = ref<string | undefined>(undefined)
    const enabled = ref(false)
    const { definitions } = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled })
    await nextTick()
    expect(mockApi.getSourceFilters).not.toHaveBeenCalled()

    // source set but still not enabled — no fetch
    sourceId.value = 's1'
    await nextTick()
    await nextTick()
    expect(mockApi.getSourceFilters).not.toHaveBeenCalled()

    // now flip to enabled — first fetch
    enabled.value = true
    await nextTick()
    await nextTick()
    expect(mockApi.getSourceFilters).toHaveBeenCalledTimes(1)
    expect(definitions.value).toEqual(DEFS)

    // switch to a new source while still enabled — one more fetch
    sourceId.value = 's2'
    await nextTick()
    await nextTick()
    expect(mockApi.getSourceFilters).toHaveBeenCalledTimes(2)

    // return to a seen source — served from cache, no refetch
    sourceId.value = 's1'
    await nextTick()
    await nextTick()
    expect(mockApi.getSourceFilters).toHaveBeenCalledTimes(2)
  })

  it('apply builds the change array as a diff vs defaults, including group children', async () => {
    const sourceId = ref<string | undefined>('s1')
    const filters = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled: ref(true) })
    await nextTick()
    await nextTick()

    filters.draft.value[filterPath(0)] = true
    filters.draft.value[filterPath(0, 1)] = 'INCLUDE'
    filters.apply()

    expect(filters.appliedChanges.value).toEqual([
      { position: 0, checkBoxState: true },
      { position: 1, groupChange: { position: 0, triState: 'INCLUDE' } },
    ])
    expect(filters.activeCount.value).toBe(2)
  })

  it('reset restores the draft to defaults without touching applied changes', async () => {
    const sourceId = ref<string | undefined>('s1')
    const filters = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled: ref(true) })
    await nextTick()
    await nextTick()

    filters.draft.value[filterPath(0)] = true
    filters.reset()
    expect(filters.draft.value[filterPath(0)]).toBe(false)
    expect(filters.appliedChanges.value).toEqual([])
  })

  it('seeds a concrete default for a sort filter with null default and emits no change when left untouched', async () => {
    const SORT_DEFS: SourceFilterDto[] = [
      { type: 'sort', position: 0, name: 'Sort', default: null, values: ['Title', 'Date'] },
    ]
    mockApi.getSourceFilters.mockResolvedValue({ success: true, data: SORT_DEFS })

    const sourceId = ref<string | undefined>('s-sort')
    const filters = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled: ref(true) })
    await nextTick()
    await nextTick()

    expect(filters.draft.value[filterPath(0)]).toEqual({ ascending: false, index: 0 })

    filters.apply()
    expect(filters.appliedChanges.value).toEqual([])
    expect(filters.activeCount.value).toBe(0)
  })

  it('persists applied changes and activeCount when the source changes within the same extension', async () => {
    const sourceId = ref<string | undefined>('s1')
    const filters = useSourceFilters({ pkgName: ref('pkg'), sourceId, enabled: ref(true) })
    await nextTick()
    await nextTick()
    filters.draft.value[filterPath(0)] = true
    filters.apply()
    expect(filters.appliedChanges.value).toHaveLength(1)
    expect(filters.activeCount.value).toBe(1)

    sourceId.value = 's2'
    await nextTick()
    await nextTick()
    expect(filters.appliedChanges.value).toHaveLength(1)
    expect(filters.activeCount.value).toBe(1)
  })
})
