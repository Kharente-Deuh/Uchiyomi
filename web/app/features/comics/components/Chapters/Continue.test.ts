// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicProgressContinue } from '../../types'
import type { Chapter } from '~/features/chapters/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import Continue from './Continue.vue'

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'c1',
    title: 'Chapter 1',
    number: 1,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    sourceChapterSlug: '1',
    pagesNb: 20,
    download: 100,
    ...overrides,
  }
}

async function mount(props: {
  comicId?: string
  continue?: ComicProgressContinue
  chapters?: Chapter[]
  sort?: 'asc' | 'desc'
} = {}): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Continue, {
      comicId: props.comicId ?? 'c1',
      continue: props.continue,
      chapters: props.chapters ?? [
        chapter({ id: 'ch-1', number: 1, pagesNb: 20 }),
        chapter({ id: 'ch-2', number: 2, pagesNb: 20 }),
      ],
      sort: props.sort ?? 'desc',
    })]),
  })
}

describe('comicsChaptersContinue', () => {
  it('renders "Start" button for the first chapter when no continue progress is present', async () => {
    const wrapper = await mount({
      continue: undefined,
      chapters: [
        chapter({ id: 'ch-1', number: 1 }),
        chapter({ id: 'ch-2', number: 2 }),
      ],
    })

    const btn = wrapper.findComponent(VBtn)
    expect(btn.props('text')).toBe('Start')
    expect(btn.props('color')).toBe('primary')
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1/ch-1')
  })

  it('renders "Continue" button for the current chapter when it is partially read', async () => {
    const wrapper = await mount({
      continue: { chapterId: 'ch-2', page: 10 },
      chapters: [
        chapter({ id: 'ch-1', number: 1 }),
        chapter({ id: 'ch-2', number: 2 }),
      ],
    })

    const btn = wrapper.findComponent(VBtn)
    expect(btn.props('text')).toBe('Continue')
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1/ch-2')
  })

  it('advances to the next chapter when the current chapter is fully read', async () => {
    const wrapper = await mount({
      continue: { chapterId: 'ch-1', page: 20 },
      chapters: [
        chapter({ id: 'ch-1', number: 1, pagesNb: 20 }),
        chapter({ id: 'ch-2', number: 2, pagesNb: 20 }),
      ],
    })

    const btn = wrapper.findComponent(VBtn)
    expect(btn.props('text')).toBe('Continue')
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1/ch-2')
  })

  it('renders "Up to date" when the last chapter is fully read', async () => {
    const wrapper = await mount({
      continue: { chapterId: 'ch-2', page: 20 },
      chapters: [
        chapter({ id: 'ch-1', number: 1, pagesNb: 20 }),
        chapter({ id: 'ch-2', number: 2, pagesNb: 20 }),
      ],
    })

    const btn = wrapper.findComponent(VBtn)
    expect(btn.props('text')).toBe('Up to date !')
    expect(btn.props('readonly')).toBe(true)
    expect(btn.props('color')).toBe('secondary')
    expect(wrapper.find('a').exists()).toBe(false)
  })

  it('handles chapters passed in asc sort order properly', async () => {
    const wrapper = await mount({
      sort: 'asc',
      continue: { chapterId: 'ch-1', page: 20 },
      chapters: [
        chapter({ id: 'ch-1', number: 1, pagesNb: 20 }),
        chapter({ id: 'ch-2', number: 2, pagesNb: 20 }),
      ],
    })

    const btn = wrapper.findComponent(VBtn)
    expect(btn.props('text')).toBe('Continue')
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1/ch-2')
  })
})
