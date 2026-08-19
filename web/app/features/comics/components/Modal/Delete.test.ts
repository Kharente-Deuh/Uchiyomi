// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Comic } from '../../types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Delete from './Delete.vue'

const { deleteById } = vi.hoisted(() => ({
  deleteById: vi.fn(),
}))

function comicsApiStub(): { deleteById: typeof deleteById } {
  return { deleteById }
}

mockNuxtImport('createComicsApi', () => comicsApiStub)

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

function comic(): Comic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: ASURA_SOURCE_NAME,
    status: 'ongoing',
    type: 'manhwa',
    author: '',
    artist: '',
    description: '',
    cover: '/cover',
    genres: [],
    altTitles: [],
    chapterCount: 1,
  }
}

async function mount(): Promise<{ wrapper: VueWrapper, deleted: ReturnType<typeof vi.fn> }> {
  const show = ref(true)
  const deleted = vi.fn()

  const wrapper = await mountSuspended(
    {
      render: () => h(VApp, () => [h(Delete, {
        'modelValue': show.value,
        'comic': comic(),
        'onUpdate:modelValue': (isShown: boolean) => {
          show.value = isShown
        },
        'onDeleted': deleted,
      })]),
    },
    { global: { stubs: { OrganismModalConfirmation: ConfirmationStub } } },
  )

  return { wrapper, deleted }
}

beforeEach(() => {
  deleteById.mockReset()
  deleteById.mockResolvedValue({ success: true, data: undefined })
})

describe('comicsModalDelete', () => {
  it('names the comic in the confirmation text', async () => {
    const { wrapper } = await mount()
    expect(wrapper.findComponent(ConfirmationStub).props('text')).toContain('Solo Leveling')
  })

  it('deletes the comic on confirmation', async () => {
    const { wrapper, deleted } = await mount()

    await wrapper.find('[data-test="confirm"]').trigger('click')

    expect(deleteById).toHaveBeenCalledWith('c1')
    expect(deleted).toHaveBeenCalled()
  })

  it('does not emit deleted when the request fails', async () => {
    deleteById.mockResolvedValue({ success: false, error: { status: 404, message: 'missing' } })
    const { wrapper, deleted } = await mount()

    await wrapper.find('[data-test="confirm"]').trigger('click')

    expect(deleted).not.toHaveBeenCalled()
  })
})
