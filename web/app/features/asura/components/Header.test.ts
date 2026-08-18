// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Header from './Header.vue'

const { search, sort, status, type, isLoading, mdAndDown } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    search: ref<string | undefined>(),
    sort: ref('popular'),
    status: ref<string | undefined>(),
    type: ref<string | undefined>(),
    isLoading: ref(false),
    mdAndDown: ref(false),
  }
})

vi.mock('../composables/asura-search.composable', () => ({
  useAsuraSearch: () => ({
    search,
    sort,
    status,
    type,
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

const SortStub = defineComponent({
  name: 'AsuraInputSort',
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
          AsuraInputSort: SortStub,
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

describe('asuraHeader', () => {
  it('renders search and filters', async () => {
    const wrapper = await mount()

    expect(wrapper.find('[data-test="search"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="sort"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="status"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="type"]').exists()).toBe(true)
  })

  it('stacks the layout on small screens', async () => {
    mdAndDown.value = true
    const wrapper = await mount()

    expect(wrapper.find('.flex-column').exists()).toBe(true)
  })
})
