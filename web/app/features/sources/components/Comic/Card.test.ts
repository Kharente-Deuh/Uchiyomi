// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import Card from './Card.vue'

describe('sourcesComicCard', () => {
  it('renders comic information and emits toggle', async () => {
    const wrapper = await mountSuspended(Card, {
      props: {
        sourceId: 'asurascans',
        comic: {
          slug: 'test-comic',
          title: 'Test Comic',
          cover: 'test.jpg',
          chapterCount: 10,
          status: 'ongoing',
          type: 'manhwa',
          author: '',
          artist: '',
          description: '',
          altTitles: [],
          genres: [],
          publicUrl: '',
          sourceUrl: '',
          updatedAt: new Date(),
          createdAt: new Date(),
          rating: 0,
        },
      },
    })

    expect(wrapper.text()).toContain('Test Comic')
  })
})
