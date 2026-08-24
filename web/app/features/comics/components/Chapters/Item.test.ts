// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Chapter } from '~/features/chapters/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VProgressCircular } from 'vuetify/components'
import Item from './Item.vue'

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'c1',
    title: 'The beginning',
    number: 12,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    sourceChapterSlug: '12',
    pagesNb: 20,
    download: 100,
    ...overrides,
  }
}

async function mount(value: Chapter = chapter(), isRetryLoading = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Item, {
      chapter: value,
      retryLoading: isRetryLoading,
      selectable: false,
      onRetry: vi.fn(),
    })]),
  })
}

describe('comicsChaptersItem', () => {
  it('renders the chapter number and title', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Chapter 12')
    expect(wrapper.text()).toContain('The beginning')
    expect(wrapper.find('.readable-chapter').exists()).toBe(true)
  })

  it('marks an early-access chapter', async () => {
    const wrapper = await mount(chapter({
      earlyAccessUntil: new Date('2099-01-01'),
    }))

    expect(wrapper.find('.early-access-chapter').exists()).toBe(true)
    expect(wrapper.text()).toContain('Unlocks')
  })

  it('shows a check when the chapter is downloaded', async () => {
    const wrapper = await mount(chapter({ download: 100 }))
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('shows progress while a past early-access chapter is downloading', async () => {
    const wrapper = await mount(chapter({
      download: 40,
      earlyAccessUntil: new Date('2020-01-01'),
    }))

    expect(wrapper.findComponent(VProgressCircular).props('modelValue')).toBe(40)
  })

  it('emits retry from the error icon', async () => {
    const onRetry = vi.fn()
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [h(Item, {
        chapter: chapter({ download: -1 }),
        retryLoading: false,
        selectable: false,
        onRetry,
      })]),
    })

    await wrapper.find('.v-icon').trigger('click')

    expect(onRetry).toHaveBeenCalled()
  })
})
