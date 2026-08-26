// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Chapter } from '~/features/chapters/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VIcon } from 'vuetify/components'
import Chapters from './index.vue'

const { getByComicId, getByIds, retryDownload, saveProgress, deleteProgress, pause, resume } = vi.hoisted(() => ({
  getByComicId: vi.fn(),
  getByIds: vi.fn(),
  retryDownload: vi.fn(),
  saveProgress: vi.fn(),
  deleteProgress: vi.fn(),
  pause: vi.fn(),
  resume: vi.fn(),
}))

function chaptersApiStub(): {
  getByComicId: typeof getByComicId
  getByIds: typeof getByIds
  retryDownload: typeof retryDownload
  saveProgress: typeof saveProgress
  deleteProgress: typeof deleteProgress
} {
  return { getByComicId, getByIds, retryDownload, saveProgress, deleteProgress }
}

function intervalStub(_fn: () => void): { pause: typeof pause, resume: typeof resume } {
  return { pause, resume }
}

mockNuxtImport('createChaptersApi', () => chaptersApiStub)
mockNuxtImport('useIntervalFn', () => intervalStub)

const ItemStub = defineComponent({
  name: 'ComicsChaptersItem',
  props: {
    chapter: { type: Object, required: true },
    retryLoading: { type: Boolean, default: false },
    selected: { type: Boolean, default: false },
    selectable: { type: Boolean, default: false },
    disabled: { type: Boolean, default: false },
  },
  emits: ['retry', 'update:selected', 'updateProgress'],
  template: `
    <div>
      <button data-test="chapter" @click="$emit('retry')">{{ chapter.number }}</button>
      <button data-test="toggle-select" @click="$emit('update:selected')">Select {{ chapter.number }}</button>
      <button data-test="mark-read" @click="$emit('updateProgress', 'read')">Read {{ chapter.number }}</button>
      <button data-test="mark-unread" @click="$emit('updateProgress', 'unread')">Unread {{ chapter.number }}</button>
    </div>
  `,
})

const ActionsStub = defineComponent({
  name: 'ComicsChaptersActions',
  props: {
    comicId: { type: String, required: true },
    modelValue: { type: Array, required: true },
  },
  emits: ['refetchChapters', 'update:modelValue'],
  template: '<div data-test="actions-component"><button data-test="action-refetch" @click="$emit(\'refetchChapters\', true)">Action</button></div>',
})

const ContinueStub = defineComponent({
  name: 'ComicsChaptersContinue',
  template: '<div data-test="continue-stub" />',
})

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'c1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: '1',
    pagesNb: 10,
    download: 100,
    ...overrides,
  }
}

async function mount(id = 'c1', onRefetchProgress = vi.fn()): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(Chapters, { id, onRefetchProgress })]) },
    {
      global: {
        stubs: {
          ComicsChaptersItem: ItemStub,
          ComicsChaptersActions: ActionsStub,
          ComicsChaptersContinue: ContinueStub,
          VVirtualScroll: false,
        },
      },
    },
  )
}

beforeEach(() => {
  getByComicId.mockReset()
  getByIds.mockReset()
  retryDownload.mockReset()
  saveProgress.mockReset()
  deleteProgress.mockReset()
  pause.mockReset()
  resume.mockReset()
  getByComicId.mockResolvedValue({ success: true, data: [] })
  getByIds.mockResolvedValue({ success: true, data: [] })
  retryDownload.mockResolvedValue({ success: true, data: undefined })
  saveProgress.mockResolvedValue({ success: true, data: { page: 10, updatedAt: new Date() } })
  deleteProgress.mockResolvedValue({ success: true, data: undefined })
})

describe('comicsChapters', () => {
  it('fetches chapters on mount', async () => {
    await mount()
    expect(getByComicId).toHaveBeenCalledWith('c1')
  })

  it('shows an empty state when there are no chapters', async () => {
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.text()).toContain('No chapters'))
  })

  it('renders a row per chapter', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-2', number: 2 }), chapter()] })
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').map(n => n.text())).toEqual(['2', '1']))
  })

  it('toggles sort between latest and oldest', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-2', number: 2 }), chapter()] })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').length).toBe(2))

    expect(wrapper.text()).toContain('Latest')

    await wrapper.findAll('button').find(b => b.text().includes('Latest'))!.trigger('click')

    expect(wrapper.text()).toContain('Oldest')
    expect(wrapper.findAll('[data-test="chapter"]').map(n => n.text())).toEqual(['1', '2'])
  })

  it('resumes polling when a download is in progress', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ download: 40 })] })
    await mount()

    await vi.waitFor(() => expect(resume).toHaveBeenCalled())
  })

  it('retries a failed download from the chapter row', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ download: -1 })] })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.find('[data-test="chapter"]').exists()).toBe(true))

    await wrapper.find('[data-test="chapter"]').trigger('click')

    expect(retryDownload).toHaveBeenCalledWith('ch-1')
  })

  it('marks a chapter as read and emits refetchProgress', async () => {
    const onRefetchProgress = vi.fn()
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-1', pagesNb: 10 })] })
    const wrapper = await mount('c1', onRefetchProgress)
    await vi.waitFor(() => expect(wrapper.find('[data-test="mark-read"]').exists()).toBe(true))

    await wrapper.find('[data-test="mark-read"]').trigger('click')

    expect(saveProgress).toHaveBeenCalledWith({ id: 'ch-1', page: 10 })
    expect(onRefetchProgress).toHaveBeenCalled()
  })

  it('marks a chapter as unread and emits refetchProgress', async () => {
    const onRefetchProgress = vi.fn()
    getByComicId.mockResolvedValue({
      success: true,
      data: [chapter({ id: 'ch-1', pagesNb: 10, progress: { page: 10, updatedAt: new Date() } })],
    })
    const wrapper = await mount('c1', onRefetchProgress)
    await vi.waitFor(() => expect(wrapper.find('[data-test="mark-unread"]').exists()).toBe(true))

    await wrapper.find('[data-test="mark-unread"]').trigger('click')

    expect(deleteProgress).toHaveBeenCalledWith('ch-1')
    expect(onRefetchProgress).toHaveBeenCalled()
  })

  it('select-all skips future early-access chapters', async () => {
    const pastEarlyAccess = new Date(Date.now() - 3_600_000)
    const futureEarlyAccess = new Date(Date.now() + 3_600_000)
    getByComicId.mockResolvedValue({
      success: true,
      data: [
        chapter({ id: 'ch-1', number: 1, earlyAccessUntil: pastEarlyAccess }),
        chapter({ id: 'ch-2', number: 2, earlyAccessUntil: futureEarlyAccess }),
        chapter({ id: 'ch-3', number: 3 }),
      ],
    })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').length).toBe(3))

    const selectAllIcon = wrapper.findAllComponents(VIcon).find(icon => icon.props('size') === 'large')!
    await selectAllIcon.trigger('click')

    expect(wrapper.find('[data-test="actions-component"]').exists()).toBe(true)
    expect(
      (wrapper.findComponent(ActionsStub).props('modelValue') as Chapter[]).map(selected => selected.id),
    ).toEqual(['ch-3', 'ch-1'])
  })

  it('selects and deselects all selectable chapters when clicking the select-all icon', async () => {
    getByComicId.mockResolvedValue({
      success: true,
      data: [
        chapter({ id: 'ch-1', number: 1 }),
        chapter({ id: 'ch-2', number: 2 }),
      ],
    })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').length).toBe(2))

    // Click select all
    const selectAllIcon = wrapper.findAllComponents(VIcon).find(icon => icon.props('size') === 'large')!
    await selectAllIcon.trigger('click')

    // Actions component should now be rendered
    expect(wrapper.find('[data-test="actions-component"]').exists()).toBe(true)

    // Click select all again to deselect
    const deselectAllIcon = wrapper.findAllComponents(VIcon).find(icon => icon.props('size') === 'large')!
    await deselectAllIcon.trigger('click')
    expect(wrapper.find('[data-test="actions-component"]').exists()).toBe(false)
  })

  it('toggles selection on an individual chapter', async () => {
    getByComicId.mockResolvedValue({
      success: true,
      data: [chapter({ id: 'ch-1', number: 1 })],
    })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.find('[data-test="toggle-select"]').exists()).toBe(true))

    await wrapper.find('[data-test="toggle-select"]').trigger('click')
    expect(wrapper.find('[data-test="actions-component"]').exists()).toBe(true)

    await wrapper.find('[data-test="toggle-select"]').trigger('click')
    expect(wrapper.find('[data-test="actions-component"]').exists()).toBe(false)
  })
})
