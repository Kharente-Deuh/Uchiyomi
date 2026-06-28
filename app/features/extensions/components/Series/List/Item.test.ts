// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '~~/shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceSearchItemDto } from '#shared/dto/catalogue/source-search.dto'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VBtn, VChip, VSkeletonLoader } from 'vuetify/components'
import Item from './Item.vue'

const baseManga: SourceSearchItemDto = {
  id: 'manga-1',
  title: 'Example Manga',
  thumbnailUrl: null,
  inLibrary: false,
  sourceUrl: null,
}

const baseSummary: MangaChapterSummaryDto = {
  chapterCount: 12,
  lastChapter: {
    name: 'Chapter 12',
    uploadedAt: new Date().toISOString(),
  },
}

function wrap(
  manga: SourceSearchItemDto,
  status?: ChapterSummaryStatus,
  summary?: MangaChapterSummaryDto,
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [
      h(Item, { manga, status, summary }),
    ]),
  }
}

describe('extensionsSeriesListItem', () => {
  it('renders the manga title', async () => {
    const wrapper = await mountSuspended(wrap(baseManga))
    expect(wrapper.text()).toContain('Example Manga')
  })

  it('renders an external link when sourceUrl is set', async () => {
    const wrapper = await mountSuspended(wrap({ ...baseManga, sourceUrl: 'https://example.com/manga' }))
    const link = wrapper.find('a[href="https://example.com/manga"]')
    expect(link.exists()).toBe(true)
  })

  it('does not render an external link when sourceUrl is null', async () => {
    const wrapper = await mountSuspended(wrap({ ...baseManga, sourceUrl: null }))
    expect(wrapper.find('a[href]').exists()).toBe(false)
  })

  it('shows chapter count and last chapter info on success with a summary', async () => {
    const wrapper = await mountSuspended(wrap(baseManga, 'success', baseSummary))
    expect(wrapper.text()).toContain('12 Chapters')
    expect(wrapper.text()).toContain('Chapter 12')
  })

  it('renders skeleton loaders when status is loading', async () => {
    const wrapper = await mountSuspended(wrap(baseManga, 'loading'))
    expect(wrapper.findComponent(VSkeletonLoader).exists()).toBe(true)
  })

  it('renders skeleton loaders when status is queued', async () => {
    const wrapper = await mountSuspended(wrap(baseManga, 'queued'))
    expect(wrapper.findComponent(VSkeletonLoader).exists()).toBe(true)
  })

  it('renders the error message when status is error', async () => {
    const wrapper = await mountSuspended(wrap(baseManga, 'error'))
    expect(wrapper.text()).toContain('Failed to load data')
  })

  it('shows the add button when the manga is not in the library', async () => {
    const wrapper = await mountSuspended(wrap({ ...baseManga, inLibrary: false }))
    expect(wrapper.findComponent(VChip).exists()).toBe(false)
    expect(wrapper.findComponent(VBtn).exists()).toBe(true)
  })

  it('shows the already-in-library chip when the manga is in the library', async () => {
    const wrapper = await mountSuspended(wrap({ ...baseManga, inLibrary: true }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.exists()).toBe(true)
    expect(wrapper.text()).toContain('Already in library')
  })
})
