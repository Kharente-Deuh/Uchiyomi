// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import TestBtn from './Test.vue'

const { test } = vi.hoisted(() => ({ test: vi.fn() }))

vi.mock('~/features/oidc/composables/oidc.composable', () => ({
  useOidc: () => ({ test, testLoading: ref(false) }),
}))

const ModalStub = defineComponent({
  name: 'OidcModalTest',
  props: { modelValue: { type: Boolean, default: false }, data: { type: Object, default: undefined } },
  template: '<div data-test="modal-stub" />',
})

async function mount(): Promise<{ wrapper: VueWrapper, setIssuerUrl: (value: string) => void }> {
  const issuerUrl = ref('')
  const wrapper = await mountSuspended({
    render: () => h(VApp, () => [h(TestBtn, { issuerUrl: issuerUrl.value })]),
  }, { global: { stubs: { OidcModalTest: ModalStub } } })

  return { wrapper, setIssuerUrl: (value: string) => {
    issuerUrl.value = value
  } }
}

function expectDisabled(wrapper: VueWrapper, isDisabled: boolean): void {
  expect(wrapper.find('button').attributes('disabled') !== undefined).toBe(isDisabled)
}

beforeEach(() => {
  test.mockReset()
})

describe('oidcBtnTest', () => {
  it('starts disabled', async () => {
    const { wrapper } = await mount()
    expectDisabled(wrapper, true)
  })

  it('enables itself once the issuer url becomes valid', async () => {
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))
  })

  it('stays disabled for a value that is not a url', async () => {
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('not-a-url')
    await vi.waitFor(() => expectDisabled(wrapper, true))
  })

  it('disables itself again when the url is emptied', async () => {
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))

    setIssuerUrl('')
    await vi.waitFor(() => expectDisabled(wrapper, true))
  })

  it('probes the current issuer url on click', async () => {
    test.mockResolvedValue({ issuer: 'https://id.example.org' })
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))
    await wrapper.find('button').trigger('click')

    expect(test).toHaveBeenCalledWith('https://id.example.org')
  })

  it('opens the result modal once the probe succeeds', async () => {
    test.mockResolvedValue({ issuer: 'https://id.example.org' })
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))
    await wrapper.find('button').trigger('click')

    await vi.waitFor(() => {
      expect(wrapper.findComponent(ModalStub).props('modelValue')).toBe(true)
      expect(wrapper.findComponent(ModalStub).props('data')).toEqual({ issuer: 'https://id.example.org' })
    })
  })

  it('leaves the modal closed when the probe fails', async () => {
    test.mockResolvedValue(undefined)
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))
    await wrapper.find('button').trigger('click')

    expect(wrapper.findComponent(ModalStub).props('modelValue')).toBe(false)
  })

  it('does not probe while disabled', async () => {
    const { wrapper } = await mount()

    await wrapper.find('button').trigger('click')

    expect(test).not.toHaveBeenCalled()
  })

  it('clears the probe result once the modal is closed', async () => {
    test.mockResolvedValue({ issuer: 'https://id.example.org' })
    const { wrapper, setIssuerUrl } = await mount()

    setIssuerUrl('https://id.example.org')
    await vi.waitFor(() => expectDisabled(wrapper, false))
    await wrapper.find('button').trigger('click')

    await vi.waitFor(() => expect(wrapper.findComponent(ModalStub).props('data')).toEqual({ issuer: 'https://id.example.org' }))

    await wrapper.findComponent(ModalStub).vm.$emit('update:modelValue', false)

    await vi.waitFor(() => expect(wrapper.findComponent(ModalStub).props('data')).toBeUndefined())
  })
})
