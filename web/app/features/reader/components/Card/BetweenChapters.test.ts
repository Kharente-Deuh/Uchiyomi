// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { AdjacentChapter, Chapter } from '~/features/chapters/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import BetweenChapters from './BetweenChapters.vue'

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-2',
    comicId: 'c1',
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: 'ch-2',
    title: 'Two',
    number: 2,
    pagesNb: 3,
    download: 100,
    ...overrides,
  }
}

function adjacent(overrides: Partial<AdjacentChapter> = {}): AdjacentChapter {
  return { id: 'ch-3', title: 'Three', number: 3, ...overrides }
}

describe('readerCardBetweenChapters', () => {
  it('shows the current and next chapter', async () => {
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(BetweenChapters, {
          comicId: 'c1',
          currentChapter: chapter(),
          nextChapter: adjacent(),
          mode: 'next',
        }),
      ]),
    })

    expect(wrapper.text()).toContain('Current chapter')
    expect(wrapper.text()).toContain('Chapter 2')
    expect(wrapper.text()).toContain('Two')
    expect(wrapper.text()).toContain('Next chapter')
    expect(wrapper.text()).toContain('Chapter 3')
    expect(wrapper.text()).toContain('Three')
  })

  it('offers an exit when there is no previous chapter', async () => {
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(BetweenChapters, {
          comicId: 'c1',
          currentChapter: chapter(),
          mode: 'previous',
        }),
      ]),
    })

    expect(wrapper.text()).toContain('There is no previous chapter')
    expect(wrapper.text()).toContain('Exit to comic')
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1')
  })
})
