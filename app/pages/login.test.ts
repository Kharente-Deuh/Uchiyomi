// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import Login from './login.vue'

const { loginMock, navigateToMock } = vi.hoisted(() => ({
  loginMock: vi.fn(),
  navigateToMock: vi.fn(),
}))

let query: Record<string, string> = {}

function useAuthMock(): ReturnType<typeof useAuth> {
  return {
    login: loginMock,
    loading: ref(false),
    user: ref(undefined),
    isAuthenticated: ref(false),
    needsAdmin: ref(false),
    minPasswordLength: ref(10),
    getSetupStatus: vi.fn(),
    fetchMe: vi.fn(),
    setup: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

function useRouteMock(): ReturnType<typeof useRoute> {
  return { query } as unknown as ReturnType<typeof useRoute>
}

mockNuxtImport('useAuth', () => useAuthMock)
mockNuxtImport('navigateTo', () => navigateToMock)
mockNuxtImport('useRoute', () => useRouteMock)

/**
 * Set field values via the internal form state and trigger form submission via
 * the real DOM submit event on AuthCard's <VForm @submit.prevent="onSubmit">.
 *
 * The page destructures `useForm` (`const { field, handleSubmit } = useForm(...)`),
 * so the field accessor is exposed directly on `$.setupState` (not under a `form`
 * binding). Accessing `$.setupState` adds no production API and needs no change to
 * login.vue — it is Vue's internal setup state, always present in tests.
 *
 * Why not DOM setValue: VTextField's onInput reads e.target.value, which
 * @vue/test-utils setValue does not reliably propagate through Vuetify's
 * proxiedModel in jsdom. The `form.trigger('submit.prevent')` exercises the real
 * AuthCard submit wiring.
 */
async function fillAndSubmit(wrapper: Awaited<ReturnType<typeof mountSuspended<typeof Login>>>, accountName: string, password: string): Promise<void> {
  const state = (wrapper.vm as any).$.setupState
  state.field('accountName').handleChange(accountName)
  state.field('password').handleChange(password)
  await flushPromises()
  await wrapper.find('form').trigger('submit.prevent')
  await flushPromises()
}

describe('login page', () => {
  afterEach(() => {
    loginMock.mockReset()
    navigateToMock.mockReset()
    query = {}
  })

  it('renders the account name + password fields and a submit button', async () => {
    const wrapper = await mountSuspended(Login)
    expect(wrapper.find('[data-test="login-accountName"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="login-password"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').exists()).toBe(true)
  })

  it('shows the expired notice from the reason query param', async () => {
    query = { reason: 'expired' }
    const wrapper = await mountSuspended(Login)
    const notice = wrapper.find('[data-test="auth-form-notice"]')
    expect(notice.exists()).toBe(true)
    expect(notice.text()).not.toBe('')
  })

  it('navigates to the redirect target on a successful login', async () => {
    query = { redirect: '/settings' }
    loginMock.mockResolvedValue({ success: true, data: undefined })
    const wrapper = await mountSuspended(Login)
    await fillAndSubmit(wrapper, 'alice', 'longenough1')
    expect(loginMock).toHaveBeenCalledWith({ accountName: 'alice', password: 'longenough1' })
    expect(navigateToMock).toHaveBeenCalledWith('/settings')
  })

  it('shows the invalid-credentials error on a 401', async () => {
    loginMock.mockResolvedValue({ success: false, error: { status: 401, message: 'x' } })
    const wrapper = await mountSuspended(Login)
    await fillAndSubmit(wrapper, 'alice', 'longenough1')
    const errorEl = wrapper.find('[data-test="auth-form-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).not.toBe('')
    // Must not show the tooMany/generic copy — this is the invalid-credentials (401) path
    expect(errorEl.text()).not.toContain('Too many')
    expect(errorEl.text()).not.toContain('tooMany')
  })
})
