// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import Setup from './setup.vue'

const { setupMock, navigateToMock } = vi.hoisted(() => ({
  setupMock: vi.fn(),
  navigateToMock: vi.fn(),
}))

function useAuthMock(): ReturnType<typeof useAuth> {
  return {
    setup: setupMock,
    login: vi.fn(),
    loading: ref(false),
    user: ref(undefined),
    isAuthenticated: ref(false),
    needsAdmin: ref(true),
    minPasswordLength: ref(10),
    getSetupStatus: vi.fn(),
    fetchMe: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

mockNuxtImport('useAuth', () => useAuthMock)
mockNuxtImport('navigateTo', () => navigateToMock)

/**
 * Set field values via the internal form state and trigger form submission via
 * the real DOM submit event on AuthCard's <VForm @submit.prevent="onSubmit">.
 *
 * The page destructures `useForm` (`const { field, handleSubmit } = useForm(...)`),
 * so the field accessor is exposed directly on `$.setupState` (not under a `form`
 * binding). Accessing `$.setupState` adds no production API and needs no change to
 * setup.vue — it is Vue's internal setup state, always present in tests.
 *
 * Why not DOM setValue: VTextField's onInput reads e.target.value, which
 * @vue/test-utils setValue does not reliably propagate through Vuetify's
 * proxiedModel in jsdom. The `form.trigger('submit.prevent')` exercises the real
 * AuthCard submit wiring.
 */
async function fillAndSubmit(
  wrapper: Awaited<ReturnType<typeof mountSuspended<typeof Setup>>>,
  accountName: string,
  displayName: string,
  password: string,
  confirmPassword: string,
): Promise<void> {
  const state = (wrapper.vm as any).$.setupState
  state.field('accountName').handleChange(accountName)
  state.field('displayName').handleChange(displayName)
  state.field('password').handleChange(password)
  state.field('confirmPassword').handleChange(confirmPassword)
  await flushPromises()
  await wrapper.find('form').trigger('submit.prevent')
  await flushPromises()
}

describe('setup page', () => {
  afterEach(() => {
    setupMock.mockReset()
    navigateToMock.mockReset()
  })

  it('renders all four fields', async () => {
    const wrapper = await mountSuspended(Setup)
    expect(wrapper.find('[data-test="setup-accountName"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="setup-displayName"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="setup-password"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="setup-confirm"]').exists()).toBe(true)
  })

  it('does not submit when the passwords differ', async () => {
    const wrapper = await mountSuspended(Setup)
    await fillAndSubmit(wrapper, 'admin', 'Admin', 'longenough1', 'different99')
    expect(setupMock).not.toHaveBeenCalled()
  })

  it('submits and redirects home on success', async () => {
    setupMock.mockResolvedValue({ success: true, data: {} })
    const wrapper = await mountSuspended(Setup)
    await fillAndSubmit(wrapper, 'admin', 'Admin', 'longenough1', 'longenough1')
    expect(setupMock).toHaveBeenCalledWith({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' })
    expect(navigateToMock).toHaveBeenCalledWith('/')
  })
})
