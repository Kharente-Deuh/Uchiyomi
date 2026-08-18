// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Confirmation from './Index.vue'

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
  name: 'OrganismModalConfirmationCard',
  props: {
    text: { type: String, required: true },
    loading: { type: Boolean, default: false },
    confirmText: { type: String, default: '' },
  },
  template: '<div data-test="card">{{ text }}<slot /></div>',
})

const stubs = {
  VDialog: DialogStub,
  VBottomSheet: SheetStub,
  OrganismModalConfirmationCard: CardStub,
}

beforeEach(() => {
  mobile.value = false
})

describe('organismModalConfirmation', () => {
  it('renders a dialog on desktop with the confirmation text', async () => {
    const wrapper = await mountSuspended(
      {
        render: () => h(VApp, () => [h(Confirmation, {
          modelValue: true,
          text: 'Are you sure?',
        }, { default: () => 'EXTRA' })]),
      },
      { global: { stubs } },
    )

    expect(wrapper.find('[data-test="dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Are you sure?')
    expect(wrapper.text()).toContain('EXTRA')
  })

  it('renders a bottom sheet on mobile', async () => {
    mobile.value = true
    const wrapper = await mountSuspended(
      {
        render: () => h(VApp, () => [h(Confirmation, { modelValue: true, text: 'x' })]),
      },
      { global: { stubs } },
    )

    expect(wrapper.find('[data-test="sheet"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="dialog"]').exists()).toBe(false)
  })
})
