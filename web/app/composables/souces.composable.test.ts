// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { ComicSource } from '~/features/comics/types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { ASURA_SCANS_URL, ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME, KING_OF_SHOJO_URL } from '~/constants'
import { useSources } from './souces.composable'

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

mockNuxtImport('useI18n', () => i18nStub)

describe('useSources', () => {
  it('returns translated details for a known source', () => {
    const { getSourceDetails } = useSources()
    const details = getSourceDetails(ASURA_SOURCE_NAME)

    expect(details.url).toBe(ASURA_SCANS_URL)
    expect(details.name).toBe('sources.asurascans.title')
    expect(details.image).toBeTruthy()
    expect(details.color).toBe('#913fe2')
  })

  it('returns details for kingofshojo', () => {
    const { getSourceDetails } = useSources()
    const details = getSourceDetails(KING_OF_SHOJO_SOURCE_NAME)

    expect(details.url).toBe(KING_OF_SHOJO_URL)
    expect(details.name).toBe('sources.kingofshojo.title')
  })

  it('throws for an unknown source', () => {
    const { getSourceDetails } = useSources()

    expect(() => getSourceDetails('unknown' as ComicSource)).toThrow('Unknown source: unknown')
  })
})
