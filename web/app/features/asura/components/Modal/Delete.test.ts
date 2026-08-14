// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraSearchItem } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Delete from './Delete.vue'

const { removeComicFromLibrary } = vi.hoisted(() => ({
  removeComicFromLibrary: vi.fn(),
}))

vi.mock('~/features/asura/composables/asura-search.composable', () => ({
  useAsuraSearch: () => ({ removeComicFromLibrary }),
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

function item(): AsuraSearchItem {
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
    latestChapters: [],
    chapterCount: 1,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    internalId: 'c1',
  }
}

async function mount(): Promise<{ wrapper: VueWrapper, comic: ReturnType<typeof ref<AsuraSearchItem | undefined>> }> {
  const show = ref(true)
  const comic = ref<AsuraSearchItem | undefined>(item())

  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(Delete, {
        'modelValue': show.value,
        'comic': comic.value,
        'onUpdate:modelValue': (isShown: boolean) => {
          show.value = isShown
        },
        'onUpdate:comic': (value: AsuraSearchItem | undefined) => {
          comic.value = value
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

describe('asuraModalDelete', () => {
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
