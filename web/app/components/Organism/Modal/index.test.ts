// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Modal from './index.vue'

const { mobile } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    mobile: ref(false),
  }
})

function displayStub(): { mobile: typeof mobile } {
  return { mobile }
}

mockNuxtImport('useDisplay', () => displayStub)

const DialogStub = defineComponent({
  name: 'VDialog',
  template: '<div data-test="dialog"><slot /></div>',
})

const SheetStub = defineComponent({
  name: 'VBottomSheet',
  template: '<div data-test="sheet"><slot /></div>',
})

const CardStub = defineComponent({
  name: 'OrganismModalCard',
  props: {
    title: { type: String, default: '' },
  },
  template: '<div data-test="card">{{ title }}<slot /><slot name="actions" /></div>',
})

const stubs = {
  VDialog: DialogStub,
  VBottomSheet: SheetStub,
  OrganismModalCard: CardStub,
}

beforeEach(() => {
  mobile.value = false
})

describe('organismModal', () => {
  it('renders a dialog on desktop', async () => {
    const wrapper = await mountSuspended(
      {
        render: () => h(VApp, () => [h(Modal, {
          modelValue: true,
          title: 'My Modal',
        }, { default: () => 'BODY' })]),
      },
      { global: { stubs } },
    )

    expect(wrapper.find('[data-test="dialog"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="sheet"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('My Modal')
    expect(wrapper.text()).toContain('BODY')
  })

  it('renders a bottom sheet on mobile', async () => {
    mobile.value = true
    const wrapper = await mountSuspended(
      {
        render: () => h(VApp, () => [h(Modal, { modelValue: true, title: 'My Modal' })]),
      },
      { global: { stubs } },
    )

    expect(wrapper.find('[data-test="sheet"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="dialog"]').exists()).toBe(false)
  })
})
