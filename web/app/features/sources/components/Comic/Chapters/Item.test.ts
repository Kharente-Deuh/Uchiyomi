// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceComicChapter } from '../../../types'
import type { ComicSource } from '~/features/comics/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VProgressCircular } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import { ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME } from '~/constants'
import Item from './Item.vue'

const { retryDownload } = vi.hoisted(() => ({
  retryDownload: vi.fn(),
}))

function chaptersApiStub(): { retryDownload: typeof retryDownload } {
  return { retryDownload }
}

mockNuxtImport('createChaptersApi', () => chaptersApiStub)

function chapter(overrides: Partial<SourceComicChapter> = {}): SourceComicChapter {
  return {
    id: 'ch-1',
    title: 'The beginning',
    number: 12,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    ...overrides,
  }
}

async function mount(
  value: SourceComicChapter = chapter(),
  source: ComicSource = ASURA_SOURCE_NAME,
  comicOriginUrl = 'https://asurascans.com/series/solo-leveling',
): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Item, {
      source,
      chapter: value,
      comicOriginUrl,
    })]),
  })
}

beforeEach(() => {
  retryDownload.mockReset()
  retryDownload.mockResolvedValue({ success: true, data: undefined })
  useToast().messages.value = []
})

describe('sourcesComicChaptersItem', () => {
  it('renders the chapter number and title', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Chapter 12')
    expect(wrapper.text()).toContain('The beginning')
    expect(wrapper.find('.readable-chapter').exists()).toBe(true)
  })

  it('links to the asura chapter url', async () => {
    const wrapper = await mount()
    expect(wrapper.find('a').attributes('href')).toBe('https://asurascans.com/series/solo-leveling/chapter/12')
  })

  it('links to the kingofshojo chapter url', async () => {
    const wrapper = await mount(
      chapter({ number: 12.5 }),
      KING_OF_SHOJO_SOURCE_NAME,
      'https://kingofshojo.com/solo-leveling',
    )

    expect(wrapper.find('a').attributes('href')).toBe('https://kingofshojo.com/solo-leveling-chapter-12-5')
  })

  it('marks an early-access chapter without a link', async () => {
    const wrapper = await mount(chapter({
      earlyAccessUntil: new Date('2099-01-01'),
    }))

    expect(wrapper.find('.early-access-chapter').exists()).toBe(true)
    expect(wrapper.text()).toContain('Unlocks')
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('shows a check when the chapter is downloaded', async () => {
    const wrapper = await mount(chapter({ internalId: 'c1', download: 100 }))
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('shows progress while a past early-access chapter is downloading', async () => {
    const wrapper = await mount(chapter({
      internalId: 'c1',
      download: 40,
      earlyAccessUntil: new Date('2020-01-01'),
    }))

    expect(wrapper.findComponent(VProgressCircular).props('modelValue')).toBe(40)
  })

  it('retries a failed download from the error icon', async () => {
    const wrapper = await mount(chapter({ internalId: 'ch-1', download: -1 }))

    await wrapper.find('.v-icon').trigger('click')

    expect(retryDownload).toHaveBeenCalledWith('ch-1')
  })

  it('toasts when retry download fails', async () => {
    retryDownload.mockResolvedValue({ success: false, error: { status: 404 } })
    const wrapper = await mount(chapter({ internalId: 'ch-1', download: -1 }))

    await wrapper.find('.v-icon').trigger('click')

    await vi.waitFor(() => expect(useToast().messages.value.map(m => m.text)).toContain('Chapter not found'))
  })
})
