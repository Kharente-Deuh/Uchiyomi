// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceComicInfos } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import StatusInfos from './StatusInfos.vue'

const { addComicInLibrary, fetchChapters } = vi.hoisted(() => ({
  addComicInLibrary: vi.fn(),
  fetchChapters: vi.fn(),
}))

vi.mock('~/features/sources/composables/sources-search.composable', () => ({
  useSourceSearch: () => ({ addComicInLibrary }),
}))

vi.mock('~/features/sources/composables/sources-chapters.composable', () => ({
  useSourceChapters: () => ({ fetchChapters }),
}))

const DeleteStub = defineComponent({
  name: 'SourcesModalDelete',
  props: {
    modelValue: { type: Boolean, default: false },
    source: { type: String, required: true },
    comic: { type: Object, required: true },
  },
  template: '<div data-test="delete-modal" />',
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

async function mount(value: SourceComicInfos = infos()): Promise<{ wrapper: VueWrapper, comic: ReturnType<typeof ref<SourceComicInfos>> }> {
  const comic = ref(value)

  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(StatusInfos, {
        'source': ASURA_SOURCE_NAME,
        'comicOriginUrl': 'https://asurascans.com/series/solo-leveling',
        'modelValue': comic.value,
        'onUpdate:modelValue': (next: SourceComicInfos) => {
          comic.value = next
        },
      })]),
    },
    { global: { stubs: { SourcesModalDelete: DeleteStub } } },
  )

  return { wrapper, comic }
}

beforeEach(() => {
  addComicInLibrary.mockReset()
  fetchChapters.mockReset()
  addComicInLibrary.mockResolvedValue('c1')
})

describe('sourcesComicStatusInfos', () => {
  it('renders status, type, author and artist', async () => {
    const { wrapper } = await mount()

    expect(wrapper.text()).toContain('Ongoing')
    expect(wrapper.text()).toContain('Manhwa')
    expect(wrapper.text()).toContain('Chugong')
    expect(wrapper.text()).toContain('Jang')
  })

  it('adds the comic to the library', async () => {
    const { wrapper, comic } = await mount()

    await wrapper.findComponent(VBtn).trigger('click')

    expect(addComicInLibrary).toHaveBeenCalled()
    await vi.waitFor(() => expect(comic.value!.internalId).toBe('c1'))
    expect(fetchChapters).toHaveBeenCalledWith('solo-leveling')
  })

  it('opens the delete modal when the comic is already in the library', async () => {
    const { wrapper } = await mount(infos({ internalId: 'c1' }))

    expect(wrapper.findComponent(VBtn).props('color')).toBe('error')
    expect(wrapper.text()).toContain('Remove from library')

    await wrapper.findComponent(VBtn).trigger('click')
    expect(addComicInLibrary).not.toHaveBeenCalled()
  })
})
