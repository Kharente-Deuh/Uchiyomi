// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { Chapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Header from './Header.vue'

function comic(overrides: Partial<Comic> = {}): Comic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: ASURA_SOURCE_NAME,
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    description: 'A hunter',
    cover: '/cover',
    genres: [],
    altTitles: [],
    chapterCount: 1,
    ...overrides,
  }
}

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

describe('readerOverlayHeader', () => {
  it('links back to the comic and to reader settings', async () => {
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(Header, { comic: comic(), chapter: chapter() }),
      ]),
    })

    expect(wrapper.text()).toContain('Solo Leveling')
    expect(wrapper.text()).toContain('Chapter 2')
    expect(wrapper.text()).toContain('Two')
    expect(wrapper.find('a[href="/comic/c1"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/settings/reader"]').exists()).toBe(true)
  })
})
