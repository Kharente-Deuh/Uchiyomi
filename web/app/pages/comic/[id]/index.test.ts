// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Comic } from '~/features/comics/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import ComicPage from './index.vue'

const { getById, refreshById, getProgress, smAndDown, navigateTo, params, query } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    getById: vi.fn(),
    refreshById: vi.fn(),
    getProgress: vi.fn(),
    smAndDown: ref(false),
    navigateTo: vi.fn(),
    params: { id: 'c1' },
    query: { from: undefined as string | undefined },
  }
})

function comicsApiStub(): { getById: typeof getById, refreshById: typeof refreshById, getProgress: typeof getProgress } {
  return { getById, refreshById, getProgress }
}

function displayStub(): { smAndDown: typeof smAndDown } {
  return { smAndDown }
}

function routeStub(): { params: typeof params, query: typeof query } {
  return { params, query }
}

mockNuxtImport('createComicsApi', () => comicsApiStub)
mockNuxtImport('useDisplay', () => displayStub)
mockNuxtImport('useRoute', () => routeStub)
mockNuxtImport('navigateTo', () => navigateTo)

const StatusStub = defineComponent({
  name: 'ComicsStatusInfos',
  props: { comic: { type: Object, required: true } },
  template: '<div data-test="status-infos">{{ comic.title }}</div>',
})

const GeneralStub = defineComponent({
  name: 'ComicsGeneralInfos',
  props: { comic: { type: Object, required: true } },
  template: '<div data-test="general-infos" />',
})

const ChaptersStub = defineComponent({
  name: 'ComicsChapters',
  props: {
    id: { type: String, required: true },
    continue: { type: Object, default: undefined },
  },
  emits: ['refetchProgress'],
  template: '<div data-test="chapters"><button data-test="refetch-progress-btn" @click="$emit(\'refetchProgress\')">{{ id }}</button></div>',
})

const DeleteStub = defineComponent({
  name: 'ComicsModalDelete',
  props: {
    modelValue: { type: Boolean, default: false },
    comic: { type: Object, required: true },
  },
  template: '<div data-test="delete-modal" />',
})

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

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(ComicPage)]) },
    {
      global: {
        stubs: {
          ComicsStatusInfos: StatusStub,
          ComicsGeneralInfos: GeneralStub,
          ComicsChapters: ChaptersStub,
          ComicsModalDelete: DeleteStub,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

function buttons(wrapper: VueWrapper): ReturnType<VueWrapper['findAllComponents']> {
  return wrapper.findAllComponents(VBtn)
}

beforeEach(() => {
  getById.mockReset()
  refreshById.mockReset()
  getProgress.mockReset()
  navigateTo.mockReset()
  smAndDown.value = false
  params.id = 'c1'
  query.from = undefined
  getById.mockResolvedValue({ success: true, data: comic() })
  refreshById.mockResolvedValue({ success: true, data: comic() })
  getProgress.mockResolvedValue({ success: true, data: {} })
})

describe('comicPage', () => {
  it('loads the comic and renders its sections', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.find('[data-test="status-infos"]').text()).toBe('Solo Leveling'))
    expect(getById).toHaveBeenCalledWith('c1')
    expect(getProgress).toHaveBeenCalledWith('c1')
    expect(wrapper.find('[data-test="general-infos"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="chapters"]').text()).toContain('c1')
  })

  it('refetches progress when ComicsChapters emits refetchProgress', async () => {
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.find('[data-test="status-infos"]').exists()).toBe(true))

    getProgress.mockClear()
    await wrapper.find('[data-test="refetch-progress-btn"]').trigger('click')

    expect(getProgress).toHaveBeenCalledWith('c1')
  })

  it('shows the refresh and remove-from-library actions', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.find('[data-test="status-infos"]').exists()).toBe(true))
    expect(buttons(wrapper).some(btn => (btn as any).props('icon') === 'fa6-solid:repeat')).toBe(true)
    expect(buttons(wrapper).some(btn => (btn as any).props('icon') === 'fa6-solid:trash' && (btn as any).props('color') === 'error')).toBe(true)
    expect(wrapper.find('[data-test="delete-modal"]').exists()).toBe(true)
  })

  it('hides the refresh action when the comic cannot be refreshed', async () => {
    getById.mockResolvedValue({ success: true, data: comic({ status: 'completed' }) })

    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.find('[data-test="status-infos"]').exists()).toBe(true))
    expect(buttons(wrapper).some(btn => (btn as any).props('icon') === 'fa6-solid:repeat')).toBe(false)
  })

  it('redirects to the library when the comic is missing', async () => {
    getById.mockResolvedValue({ success: false, error: { status: 404, message: 'missing' } })

    await mount()

    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith({ name: 'library' }))
  })

  it('redirects to the feed when opened from the feed and the comic is missing', async () => {
    query.from = 'feed'
    getById.mockResolvedValue({ success: false, error: { status: 404, message: 'missing' } })

    await mount()

    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith({ name: 'feed' }))
  })
})
