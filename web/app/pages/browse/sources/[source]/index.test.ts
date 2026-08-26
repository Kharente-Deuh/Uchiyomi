// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import SourceBrowsePage from './index.vue'

vi.mock('~/features/sources/composables/sources-search.composable', () => ({
  useSourceSearch: () => ({
    isLoading: ref(false),
    series: ref([]),
    page: ref(1),
    hasNextPage: ref(false),
    addComicInLibrary: vi.fn(),
    addComicInLibraryLoading: ref({}),
    resetFilters: vi.fn(),
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
  return { params: { source: 'asurascans' }, name: 'browse-sources-source' }
}

mockNuxtImport('useRoute', () => routeStub)

describe('browse Source Page', () => {
  it('mounts properly for a valid source', async () => {
    const wrapper = await mountSuspended(SourceBrowsePage)
    expect(wrapper.exists()).toBe(true)
  })
})
