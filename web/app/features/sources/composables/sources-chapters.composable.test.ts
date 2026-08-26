// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSourceChapters } from './sources-chapters.composable'

const { getSeriesChapters } = vi.hoisted(() => ({
  getSeriesChapters: vi.fn(),
}))

vi.mock('./sources.api', () => ({
  createSourceApi: () => ({ getSeriesChapters }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

mockNuxtImport('useI18n', () => i18nStub)

describe('useSourceChapters', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes chapters composable', () => {
    const composable = useSourceChapters('asurascans')
    expect(composable.chapters.value).toEqual([])
    expect(composable.loading.value).toBe(false)
  })
})
