// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSourceSearch } from './sources-search.composable'

const { search, smAndDown } = vi.hoisted(() => ({
  search: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('./sources.api', () => ({
  createSourceApi: () => ({ search }),
}))

vi.mock('~/features/comics/composables/comics.api', () => ({
  createComicsApi: () => ({ create: vi.fn(), deleteById: vi.fn() }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function displayStub(): { smAndDown: { value: boolean } } {
  return { smAndDown }
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('useDisplay', () => displayStub)

describe('useSourceSearch', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes search composable with store bindings', () => {
    const composable = useSourceSearch('asurascans', { doSearch: false })
    expect(composable.search.value).toBeUndefined()
    expect(composable.page.value).toBe(1)
    expect(composable.isLoading.value).toBe(false)
  })
})
