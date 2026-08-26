// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { KingOfShojoComicChapter } from '~/features/kingofshojo/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VProgressCircular } from 'vuetify/components'
import Item from './Item.vue'

const { retryDownload, retryDownloadLoading } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    retryDownload: vi.fn(),
    retryDownloadLoading: ref(false),
  }
})

vi.mock('~/features/kingofshojo/composables/kingofshojo-chapters.composable', () => ({
  useKingOfShojoChapters: () => ({
    retryDownload,
    retryDownloadLoading,
  }),
}))

function chapter(overrides: Partial<KingOfShojoComicChapter> = {}): KingOfShojoComicChapter {
  return {
    id: 'ch-1',
    title: 'The beginning',
    number: 12,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    ...overrides,
  }
}

async function mount(value: KingOfShojoComicChapter = chapter()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Item, {
      chapter: value,
      comicOriginUrl: 'https://kingofshojo.com/solo-leveling',
    })]),
  })
}

beforeEach(() => {
  retryDownload.mockReset()
  retryDownloadLoading.value = false
})

describe('asuraScansComicChaptersItem', () => {
  it('links a readable chapter to the source chapter page', async () => {
    const wrapper = await mount()
    expect(wrapper.find('a').attributes('href')).toBe('https://kingofshojo.com/solo-leveling-chapter-12')
    expect(wrapper.text()).toContain('Chapter 12')
    expect(wrapper.text()).toContain('The beginning')
  })

  it('builds the chapter path from site origin, comic slug and number', async () => {
    const wrapper = await mount(chapter({ number: 111.5 }))
    expect(wrapper.find('a').attributes('href')).toBe('https://kingofshojo.com/solo-leveling-chapter-111-5')
  })

  it('shows a check when the chapter is downloaded', async () => {
    const wrapper = await mount(chapter({
      internalId: 'internal-1',
      download: 100,
    }))

    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('shows progress while a past early-access chapter is downloading', async () => {
    const wrapper = await mount(chapter({
      internalId: 'internal-1',
      download: 40,
      earlyAccessUntil: new Date('2020-01-01'),
    }))

    expect(wrapper.findComponent(VProgressCircular).props('modelValue')).toBe(40)
  })

  it('retries a failed download from the error icon', async () => {
    const wrapper = await mount(chapter({
      internalId: 'internal-1',
      download: -1,
    }))

    await wrapper.find('.v-icon').trigger('click')

    expect(retryDownload).toHaveBeenCalledWith('internal-1')
  })
})
