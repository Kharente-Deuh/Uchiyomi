// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { Chapter } from '~/features/chapters/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VProgressCircular } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import ChaptersMenu from './ChaptersMenu.vue'

const { getByComicId, getByIds } = vi.hoisted(() => ({
  getByComicId: vi.fn(),
  getByIds: vi.fn(),
}))

function createChaptersApiStub(): { getByComicId: typeof getByComicId, getByIds: typeof getByIds } {
  return { getByComicId, getByIds }
}

mockNuxtImport('createChaptersApi', () => createChaptersApiStub)

const MenuStub = defineComponent({
  name: 'VMenu',
  setup(_, { slots }) {
    return () => slots.default?.()
  },
})

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-2',
    comicId: 'c1',
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: 'ch-2',
    title: 'Two',
    number: 2,
    pagesNb: 3,
    download: 100,
    ...overrides,
  }
}

async function mount(current = chapter()): Promise<ReturnType<typeof mountSuspended>> {
  return mountSuspended({
    render: () => h(VApp, () => [
      h(ChaptersMenu, { comicId: 'c1', currentChapter: current }),
    ]),
  }, {
    global: {
      stubs: { VMenu: MenuStub },
    },
  })
}

beforeEach(() => {
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getByComicId.mockReset()
  getByIds.mockReset()
  getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-1', number: 1, title: 'One' }), chapter()] })
})

describe('readerOverlayFooterChaptersMenu', () => {
  it('loads chapters for the comic', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(getByComicId).toHaveBeenCalledWith('c1'))
    expect(wrapper.text()).toContain('Chapter 2')
  })

  it('lists chapters for the comic', async () => {
    const wrapper = await mount()
    await vi.waitFor(() => expect(getByComicId).toHaveBeenCalled())

    await vi.waitFor(() => expect(wrapper.text()).toContain('One'))
    expect(wrapper.find('a[href="/comic/c1/ch-1"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/comic/c1/ch-2"]').exists()).toBe(true)
  })

  it('shows a progress ring for an in-progress download in the list', async () => {
    getByComicId.mockResolvedValue({
      success: true,
      data: [chapter({ download: 40 })],
    })
    const wrapper = await mount(chapter({ download: 40 }))
    await vi.waitFor(() => expect(getByComicId).toHaveBeenCalled())

    await vi.waitFor(() => expect(wrapper.findComponent(VProgressCircular).exists()).toBe(true))
  })
})
