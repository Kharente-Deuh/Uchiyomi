// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import { DEFAULT_PAGE } from '~/constants'
import SetupPage from './setup.vue'

const { doSetup, navigateTo } = vi.hoisted(() => ({
  doSetup: vi.fn(),
  navigateTo: vi.fn(),
}))

function setupApiStub(): { doSetup: typeof doSetup } {
  return { doSetup }
}

mockNuxtImport('createSetupApi', () => setupApiStub)
mockNuxtImport('navigateTo', () => navigateTo)

async function mount(): Promise<VueWrapper> {
  return mountSuspended({ render: () => h(VApp, () => [h(SetupPage)]) })
}

async function fill(wrapper: VueWrapper): Promise<void> {
  await wrapper.find('[data-test="setup-username"] input').setValue('alice')
  await wrapper.find('[data-test="setup-password"] input').setValue('password12')
  await wrapper.find('[data-test="setup-confirm"] input').setValue('password12')
}

beforeEach(() => {
  doSetup.mockReset()
  navigateTo.mockReset()
  doSetup.mockResolvedValue({ success: true, data: undefined })
})

describe('setupPage', () => {
  it('renders the setup form', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Let\'s begin')
    expect(wrapper.find('[data-test="setup-username"]').exists()).toBe(true)
  })

  it('creates the account and goes to the default page', async () => {
    const wrapper = await mount()
    await fill(wrapper)
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(doSetup).toHaveBeenCalledWith({
      username: 'alice',
      password: 'password12',
    }))
    expect(navigateTo).toHaveBeenCalledWith(DEFAULT_PAGE)
  })

  it('redirects to login when setup is already closed', async () => {
    doSetup.mockResolvedValue({ success: false, error: { status: 409 } })
    const wrapper = await mount()
    await fill(wrapper)
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith({ path: '/login', query: { reason: 'setupClosed' } }))
  })

  it('shows a generic error on other failures', async () => {
    doSetup.mockResolvedValue({ success: false, error: { status: 500 } })
    const wrapper = await mount()
    await fill(wrapper)
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(wrapper.find('[data-test="auth-form-error"]').text()).toContain('Could not create the account'))
    expect(navigateTo).not.toHaveBeenCalled()
  })
})
