// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { FeedItem } from '~/features/feed/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import FeedPage from './feed.vue'

const { items, isLoading, maxPage, smAndDown, page, source, type } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    items: ref<FeedItem[]>([]),
    isLoading: ref(false),
    maxPage: ref(1),
    smAndDown: ref(false),
    page: ref(1),
    source: ref<string | undefined>(),
    type: ref<string | undefined>(),
  }
})

vi.mock('~/features/feed/composables/feed.composable', () => ({
  useFeed: () => ({
    isLoading,
    items,
    page,
    maxPage,
    source,
    type,
  }),
}))

vi.mock('@vueuse/core', async importOriginal => ({
  ...(await importOriginal<typeof import('@vueuse/core')>()),
  useIntersectionObserver: vi.fn(),
}))

function displayStub(): { smAndDown: typeof smAndDown } {
  return { smAndDown }
}

mockNuxtImport('useDisplay', () => displayStub)

const CardStub = defineComponent({
  name: 'FeedComicCard',
  props: { item: { type: Object, required: true } },
  template: '<div data-test="feed-card">{{ item.title }}</div>',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(FeedPage)]) },
    {
      global: {
        stubs: {
          FeedComicCard: CardStub,
          ComicsInputSource: true,
          ComicsInputType: true,
          MoleculePaginationFooter: true,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

function item(): FeedItem {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    cover: '/cover',
    source: 'asurascans',
    status: 'ongoing',
    type: 'manhwa',
    latestChapters: [],
  }
}

beforeEach(() => {
  items.value = []
  isLoading.value = false
  maxPage.value = 1
  smAndDown.value = false
  page.value = 1
  source.value = undefined
  type.value = undefined
})

describe('feedPage', () => {
  it('shows an empty state when there are no items', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No results')
  })

  it('renders a card per item', async () => {
    items.value = [item()]
    const wrapper = await mount()
    expect(wrapper.findAll('[data-test="feed-card"]').map(n => n.text())).toEqual(['Solo Leveling'])
  })
})
