// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSlider } from 'vuetify/components'
import Rail from './Rail.vue'

function chapter(overrides: Partial<DetailedChapter> = {}): DetailedChapter {
  return {
    id: 'ch-2',
    comicId: 'c1',
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: 'ch-2',
    title: 'Two',
    number: 2,
    pagesNb: 5,
    download: 100,
    pageUrls: ['/p1', '/p2', '/p3', '/p4', '/p5'],
    ...overrides,
  }
}

async function mount(
  page = 2,
  value: DetailedChapter = chapter(),
  isDoublePage = false,
): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [
      h(Rail, {
        'chapter': value,
        'page': page,
        'doublePage': isDoublePage,
        'onUpdate:page': () => {},
      }),
    ]),
  })
}

describe('readerOverlayRail', () => {
  it('shows the 1-based current page and the page count', async () => {
    const wrapper = await mount(2)

    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('5')
  })

  it('renders a readonly vertical reversed slider snapped to pages', async () => {
    const wrapper = await mount(2)
    const slider = wrapper.findComponent(VSlider)

    expect(slider.exists()).toBe(true)
    expect(slider.props('direction')).toBe('vertical')
    expect(slider.props('reverse')).toBe(true)
    expect(slider.props('readonly')).toBe(true)
    expect(slider.props('step')).toBe(1)
    expect(slider.props('min')).toBe(0)
    expect(slider.props('max')).toBe(4)
    expect(slider.props('modelValue')).toBe(2)
    expect(slider.props('showTicks')).toBe('always')
    expect(slider.props('color')).toBe('primary')
  })

  it('keeps max at 0 for a one-page chapter', async () => {
    const wrapper = await mount(0, chapter({
      pagesNb: 1,
      pageUrls: ['/p1'],
    }))
    const slider = wrapper.findComponent(VSlider)

    expect(wrapper.text()).toContain('1')
    expect(slider.props('max')).toBe(0)
    expect(slider.props('modelValue')).toBe(0)
  })

  it('shows a page pair and steps by two when double page is on', async () => {
    const wrapper = await mount(0, chapter(), true)
    const slider = wrapper.findComponent(VSlider)

    expect(wrapper.get('[data-test="rail-current"]').text()).toBe('1–2')
    expect(slider.props('step')).toBe(2)
    expect(slider.props('max')).toBe(4)
  })

  it('shows the leftover last page alone in double page', async () => {
    const wrapper = await mount(4, chapter(), true)

    expect(wrapper.get('[data-test="rail-current"]').text()).toBe('5')
  })
})
