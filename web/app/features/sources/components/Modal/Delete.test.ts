// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceComicInfos, SourceSearchItem } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Delete from './Delete.vue'

const { removeComicFromLibrary } = vi.hoisted(() => ({
  removeComicFromLibrary: vi.fn(),
}))

vi.mock('~/features/sources/composables/sources-search.composable', () => ({
  useSourceSearch: () => ({ removeComicFromLibrary }),
}))

const ConfirmationStub = defineComponent({
  name: 'OrganismModalConfirmation',
  props: {
    modelValue: { type: Boolean, default: false },
    text: { type: String, default: '' },
    loading: { type: Boolean, default: false },
  },
  emits: ['confirm'],
  template: '<button data-test="confirm" @click="$emit(\'confirm\')" />',
})

function item(): SourceSearchItem {
  return {
    slug: 'solo-leveling',
    title: 'Solo Leveling',
    cover: '/cover',
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: '',
    artist: '',
    description: '',
    altTitles: [],
    genres: [],
    chapterCount: 1,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    internalId: 'c1',
  }
}

async function mount(): Promise<{ wrapper: VueWrapper, comic: ReturnType<typeof ref<SourceSearchItem | undefined>> }> {
  const show = ref(true)
  const comic = ref<SourceSearchItem | undefined>(item())

  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(Delete, {
        'source': 'asurascans',
        'modelValue': show.value,
        'comic': comic.value,
        'onUpdate:modelValue': (isShown: boolean) => {
          show.value = isShown
        },
        'onUpdate:comic': (value: SourceSearchItem | SourceComicInfos | undefined) => {
          comic.value = value as SourceSearchItem
        },
      })]),
    },
    { global: { stubs: { OrganismModalConfirmation: ConfirmationStub } } },
  )

  return { wrapper, comic }
}

beforeEach(() => {
  removeComicFromLibrary.mockReset()
  removeComicFromLibrary.mockResolvedValue(true)
})

describe('sourcesModalDelete', () => {
  it('names the comic in the confirmation text', async () => {
    const { wrapper } = await mount()
    expect(wrapper.findComponent(ConfirmationStub).props('text')).toContain('Solo Leveling')
  })

  it('removes the comic on confirmation', async () => {
    const { wrapper } = await mount()

    await wrapper.find('[data-test="confirm"]').trigger('click')

    expect(removeComicFromLibrary).toHaveBeenCalled()
  })
})
