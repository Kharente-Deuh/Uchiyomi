// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Header from './Header.vue'

const { search, sort, status, type, source, isLoading, mdAndDown } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    search: ref<string | undefined>(),
    sort: ref('title'),
    status: ref<string | undefined>(),
    type: ref<string | undefined>(),
    source: ref<string | undefined>(),
    isLoading: ref(false),
    mdAndDown: ref(false),
  }
})

vi.mock('../composables/library-search.composable', () => ({
  useLibrarySearch: () => ({
    search,
    sort,
    status,
    type,
    source,
    isLoading,
  }),
}))

function displayStub(): { mdAndDown: typeof mdAndDown } {
  return { mdAndDown }
}

mockNuxtImport('useDisplay', () => displayStub)

const SearchStub = defineComponent({
  name: 'AtomInputSearch',
  template: '<input data-test="search" />',
})

const SourceStub = defineComponent({
  name: 'ComicsInputSource',
  template: '<div data-test="source" />',
})

const SortStub = defineComponent({
  name: 'LibraryInputSort',
  template: '<div data-test="sort" />',
})

const StatusStub = defineComponent({
  name: 'ComicsInputStatus',
  template: '<div data-test="status" />',
})

const TypeStub = defineComponent({
  name: 'ComicsInputType',
  template: '<div data-test="type" />',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(Header)]) },
    {
      global: {
        stubs: {
          AtomInputSearch: SearchStub,
          ComicsInputSource: SourceStub,
          LibraryInputSort: SortStub,
          ComicsInputStatus: StatusStub,
          ComicsInputType: TypeStub,
        },
      },
    },
  )
}

beforeEach(() => {
  mdAndDown.value = false
  isLoading.value = false
})

describe('libraryHeader', () => {
  it('renders search and filters on desktop', async () => {
    const wrapper = await mount()

    expect(wrapper.find('[data-test="search"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="source"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="sort"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="status"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="type"]').exists()).toBe(true)
  })

  it('collapses filters on small screens until the chevron is clicked', async () => {
    mdAndDown.value = true
    const wrapper = await mount()
    const filters = wrapper.get('.flex-wrap')

    expect(filters.attributes('style')).toContain('display: none')

    await wrapper.find('.v-icon').trigger('click')

    expect(filters.attributes('style') ?? '').not.toContain('display: none')
  })
})
