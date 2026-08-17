// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { LightComic } from '~/features/comics/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import LibraryPage from './library.vue'

const { comics, isLoading, maxPage, smAndDown, page } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    comics: ref<LightComic[]>([]),
    isLoading: ref(false),
    maxPage: ref(1),
    smAndDown: ref(false),
    page: ref(1),
  }
})

vi.mock('~/features/library/composables/library-search.composable', () => ({
  useLibrarySearch: () => ({
    isLoading,
    comics,
    page,
    maxPage,
    resetFilters: vi.fn(),
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

const HeaderStub = defineComponent({
  name: 'LibraryHeader',
  template: '<div data-test="library-header" />',
})

const CardStub = defineComponent({
  name: 'LibraryComicCard',
  props: { comic: { type: Object, required: true } },
  template: '<div data-test="comic-card">{{ comic.title }}</div>',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(LibraryPage)]) },
    {
      global: {
        stubs: {
          LibraryHeader: HeaderStub,
          LibraryComicCard: CardStub,
          MoleculePaginationFooter: true,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

function item(): LightComic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: 'asurascans',
    status: 'ongoing',
    chapterCount: 12,
    cover: '/cover',
  }
}

beforeEach(() => {
  comics.value = []
  isLoading.value = false
  maxPage.value = 1
  smAndDown.value = false
  page.value = 1
})

describe('libraryPage', () => {
  it('shows an empty state when there are no comics', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No results')
  })

  it('renders a card per comic', async () => {
    comics.value = [item()]
    const wrapper = await mount()
    expect(wrapper.findAll('[data-test="comic-card"]').map(n => n.text())).toEqual(['Solo Leveling'])
  })
})
