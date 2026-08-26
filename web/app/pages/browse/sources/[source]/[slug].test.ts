// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceComicInfos } from '~/features/sources/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import { ASURA_SOURCE_NAME } from '~/constants'
import SourceComicPage from './[slug].vue'

const { getInfosBySlug, smAndDown, navigateTo, params } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    getInfosBySlug: vi.fn(),
    smAndDown: ref(false),
    navigateTo: vi.fn(),
    params: { source: 'asurascans', slug: 'solo-leveling' },
  }
})

function sourceApiStub(): { getInfosBySlug: typeof getInfosBySlug } {
  return { getInfosBySlug }
}

function displayStub(): { smAndDown: typeof smAndDown } {
  return { smAndDown }
}

function routeStub(): { params: typeof params } {
  return { params }
}

mockNuxtImport('createSourceApi', () => sourceApiStub)
mockNuxtImport('useDisplay', () => displayStub)
mockNuxtImport('useRoute', () => routeStub)
mockNuxtImport('navigateTo', () => navigateTo)

const StatusStub = defineComponent({
  name: 'SourcesComicStatusInfos',
  props: { source: { type: String, required: true }, comicOriginUrl: { type: String, required: true } },
  template: '<div data-test="status-infos" />',
})

const GeneralStub = defineComponent({
  name: 'SourcesComicGeneralInfos',
  props: { comic: { type: Object, required: true } },
  template: '<div data-test="general-infos">{{ comic.title }}</div>',
})

const ChaptersStub = defineComponent({
  name: 'SourcesComicChapters',
  props: {
    source: { type: String, required: true },
    slug: { type: String, required: true },
    comicOriginUrl: { type: String, required: true },
  },
  template: '<div data-test="chapters">{{ slug }}</div>',
})

function infos(overrides: Partial<SourceComicInfos> = {}): SourceComicInfos {
  return {
    slug: 'solo-leveling',
    title: 'Solo Leveling',
    description: 'A hunter',
    cover: '/cover',
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    sourceUrl: '/series/solo-leveling',
    publicUrl: '/series/solo-leveling',
    altTitles: [],
    genres: [],
    chapterCount: 1,
    rating: 0,
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...overrides,
  }
}

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(SourceComicPage)]) },
    {
      global: {
        stubs: {
          SourcesComicStatusInfos: StatusStub,
          SourcesComicGeneralInfos: GeneralStub,
          SourcesComicChapters: ChaptersStub,
          OrganismPageLayout: false,
        },
      },
    },
  )
}

beforeEach(() => {
  getInfosBySlug.mockReset()
  navigateTo.mockReset()
  smAndDown.value = false
  params.source = ASURA_SOURCE_NAME
  params.slug = 'solo-leveling'
  getInfosBySlug.mockResolvedValue({ success: true, data: infos() })
  useToast().messages.value = []
})

describe('browse Source Comic Page', () => {
  it('loads the comic and renders its sections', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.find('[data-test="general-infos"]').text()).toBe('Solo Leveling'))
    expect(getInfosBySlug).toHaveBeenCalledWith('solo-leveling')
    expect(wrapper.find('[data-test="status-infos"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="chapters"]').text()).toContain('solo-leveling')
  })

  it('redirects to the source browse page when the comic is missing', async () => {
    getInfosBySlug.mockResolvedValue({ success: false, error: { status: 404, message: 'missing' } })

    await mount()

    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith(`/browse/sources/${ASURA_SOURCE_NAME}`))
    expect(useToast().messages.value.map(m => m.text)).toContain('Comic not found')
  })

  it('redirects when the source is unknown', async () => {
    params.source = 'unknown'

    await expect(mount()).rejects.toBeTruthy()
    expect(navigateTo).toHaveBeenCalledWith('/browse/sources', { replace: true })
  })
})
