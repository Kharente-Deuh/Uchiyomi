// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import Account from './index.vue'
import UpdatePassword from './Modal/UpdatePassword.vue'

const user: UserDto = {
  id: 'u1',
  accountName: 'admin',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
}

const { updateDisplayNameMock } = vi.hoisted(() => ({ updateDisplayNameMock: vi.fn() }))

function useAuthMock(): ReturnType<typeof useAuth> {
  return {
    user: ref(user),
    loading: ref(false),
    minPasswordLength: ref(10),
    isAuthenticated: ref(true),
    needsAdmin: ref(false),
    getSetupStatus: vi.fn(),
    fetchMe: vi.fn(),
    login: vi.fn(),
    setup: vi.fn(),
    updateDisplayName: updateDisplayNameMock,
    changePassword: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

mockNuxtImport('useAuth', () => useAuthMock)

describe('settingsAccount', () => {
  afterEach(() => {
    updateDisplayNameMock.mockReset()
    vi.useRealTimers()
  })

  it('renders the current displayName and account name values', async () => {
    const wrapper = await mountSuspended(Account)
    const values = wrapper.findAll('input').map(i => (i.element as HTMLInputElement).value)
    expect(values).toContain('Admin')
    expect(values).toContain('admin')
  })

  it('wires a change-password trigger and mounts the modal closed', async () => {
    const wrapper = await mountSuspended(Account)
    expect(wrapper.find('button').exists()).toBe(true)
    expect(wrapper.findComponent(UpdatePassword).props('modelValue')).toBe(false)
  })

  it('auto-saves a changed displayName after the 500ms debounce', async () => {
    vi.useFakeTimers()
    updateDisplayNameMock.mockResolvedValue({ success: true, data: { ...user, displayName: 'Renamed' } })
    const wrapper = await mountSuspended(Account)
    // Drive the debounced save directly (Vuetify input value does not propagate
    // reliably through jsdom — see login.test.ts).
    ;(wrapper.vm as any).$.setupState.saveDisplayName('Renamed')
    await vi.advanceTimersByTimeAsync(500)
    expect(updateDisplayNameMock).toHaveBeenCalledWith({ displayName: 'Renamed' })
  })

  it('does not save when the value is unchanged', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(Account)
    ;(wrapper.vm as any).$.setupState.saveDisplayName('Admin')
    await vi.advanceTimersByTimeAsync(500)
    expect(updateDisplayNameMock).not.toHaveBeenCalled()
  })

  it('reverts the field when the save fails', async () => {
    vi.useFakeTimers()
    updateDisplayNameMock.mockResolvedValue({ success: false, error: { status: 400, message: 'x' } })
    const wrapper = await mountSuspended(Account)
    const state = (wrapper.vm as any).$.setupState
    state.displayName = 'Renamed'
    state.saveDisplayName('Renamed')
    await vi.advanceTimersByTimeAsync(500)
    expect(updateDisplayNameMock).toHaveBeenCalledWith({ displayName: 'Renamed' })
    expect(state.displayName).toBe('Admin')
  })
})
