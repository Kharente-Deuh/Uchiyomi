// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Comic } from '../types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import StatusInfos from './StatusInfos.vue'

function comicsApiStub(): { refreshById: ReturnType<typeof vi.fn> } {
  return { refreshById: vi.fn() }
}

mockNuxtImport('createComicsApi', () => comicsApiStub)

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
    genres: ['action'],
    altTitles: [],
    chapterCount: 1,
    ...overrides,
  }
}

async function mount(value: Comic = comic()): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(StatusInfos, { modelValue: value })]) },
    { global: { stubs: { ComicsModalDelete: DeleteStub } } },
  )
}

describe('comicsStatusInfos', () => {
  it('renders status, type, author and artist', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Ongoing')
    expect(wrapper.text()).toContain('Manhwa')
    expect(wrapper.text()).toContain('Chugong')
    expect(wrapper.text()).toContain('Jang')
    expect(wrapper.find('[data-test="delete-modal"]').exists()).toBe(true)
  })

  it('shows the remove-from-library action', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Remove from library')
    expect(wrapper.findAllComponents(VBtn).some(btn => btn.props('color') === 'error')).toBe(true)
  })
})
