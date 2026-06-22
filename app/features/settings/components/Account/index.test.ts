// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import Account from './index.vue'
import UpdatePassword from './Modal/UpdatePassword.vue'

const user: UserDto = {
  id: 'u1',
  email: 'admin@uchiyomi.test',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
}

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
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

mockNuxtImport('useAuth', () => useAuthMock)

describe('settingsAccount', () => {
  it('renders the current displayName and email values', async () => {
    const wrapper = await mountSuspended(Account)
    const values = wrapper.findAll('input').map(i => (i.element as HTMLInputElement).value)
    expect(values).toContain('Admin')
    expect(values).toContain('admin@uchiyomi.test')
  })

  it('wires a change-password trigger and mounts the modal closed', async () => {
    // The actual VDialog open is Vuetify overlay behaviour (unreliable to drive in
    // jsdom); we assert the trigger exists and the modal starts closed.
    const wrapper = await mountSuspended(Account)
    expect(wrapper.find('button').exists()).toBe(true)
    expect(wrapper.findComponent(UpdatePassword).props('modelValue')).toBe(false)
  })
})
