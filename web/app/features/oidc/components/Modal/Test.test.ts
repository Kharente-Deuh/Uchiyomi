// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { TestResponse } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import TestModal from './Test.vue'

const ModalStub = defineComponent({
  name: 'OrganismModal',
  template: '<div><slot /></div>',
})

const response = {
  issuer: 'https://id.example.org',
  authorizationEndpoint: 'https://id.example.org/authorize',
  tokenEndpoint: 'https://id.example.org/token',
  userInfoEndpoint: 'https://id.example.org/userinfo',
  endSessionEndpoint: 'https://id.example.org/logout',
  redirectUri: 'https://app.example.org/callback',
  supportsRpInitiatedLogout: true,
} satisfies TestResponse

async function mount(data: TestResponse | undefined): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(TestModal, { modelValue: true, data })]) },
    { global: { stubs: { OrganismModal: ModalStub } } },
  )
}

function values(wrapper: VueWrapper): string[] {
  return wrapper.findAll('input[type="text"]').map(i => (i.element as HTMLInputElement).value)
}

describe('oidcModalTest', () => {
  it('renders every string field of the probe result', async () => {
    const wrapper = await mount(response)

    expect(values(wrapper)).toEqual([
      response.issuer,
      response.authorizationEndpoint,
      response.tokenEndpoint,
      response.userInfoEndpoint,
      response.endSessionEndpoint,
      response.redirectUri,
    ])
  })

  it('renders the boolean field as a checked box', async () => {
    const wrapper = await mount(response)
    const checkboxes = wrapper.findAll('input[type="checkbox"]')

    expect(checkboxes).toHaveLength(1)
    expect(checkboxes[0]!.element).toHaveProperty('checked', true)
  })

  it('renders an unchecked box when logout is unsupported', async () => {
    const wrapper = await mount({ ...response, supportsRpInitiatedLogout: false })
    expect(wrapper.find('input[type="checkbox"]').element).toHaveProperty('checked', false)
  })

  it('labels each field with its translation', async () => {
    const wrapper = await mount(response)
    expect(wrapper.text()).toContain('Authorization endpoint')
  })

  it('keeps every field read-only', async () => {
    const wrapper = await mount(response)
    expect(wrapper.findAll('input[type="text"]').every(i => i.attributes('readonly') !== undefined)).toBe(true)
  })

  it('renders no field without a probe result', async () => {
    const wrapper = await mount(undefined)
    expect(wrapper.findAll('input')).toHaveLength(0)
  })
})
