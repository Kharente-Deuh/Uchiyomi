// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { LightComic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import ComicCard from './ComicCard.vue'

function item(): LightComic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: 'asurascans',
    status: 'ongoing',
    chapterCount: 12,
    cover: '/cover/solo-leveling',
  }
}

async function mount(comic: LightComic = item()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(ComicCard, { comic })]),
  })
}

describe('libraryComicCard', () => {
  it('renders the title', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('Solo Leveling')
  })

  it('links to the library comic page', async () => {
    const wrapper = await mount()
    expect(wrapper.find('a').attributes('href')).toBe('/comic/c1?from=library')
  })

  it('shows the chapter count', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('12 Chapters')
  })
})
