// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UserDto } from '#shared/dto/identity'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { VSwitch } from 'vuetify/components'
import SensitiveContent from './SensitiveContent.vue'

const user: UserDto = {
  id: 'u1',
  accountName: 'viewer',
  displayName: 'Viewer',
  role: 'USER',
  status: 'ACTIVE',
  canManageExtensions: false,
  canDownload: false,
  allowNsfw: true,
  showNsfw: false,
}

const updateShowNsfwMock = vi.fn()

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
    updateDisplayName: vi.fn(),
    updateShowNsfw: updateShowNsfwMock,
    changePassword: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

mockNuxtImport('useAuth', () => useAuthMock)

describe('settingsSensitiveContent', () => {
  it('renders a switch reflecting the user showNsfw value', async () => {
    const wrapper = await mountSuspended(SensitiveContent)
    const sw = wrapper.findComponent(VSwitch)
    expect(sw.exists()).toBe(true)
    // showNsfw is false for the test user
    expect(sw.props('modelValue')).toBe(false)
  })

  it('renders a non-empty translated card title', async () => {
    const wrapper = await mountSuspended(SensitiveContent)
    // i18n is active in the nuxt test env and resolves the 'settings.sensitiveContent.title' key
    const text = wrapper.text()
    expect(text.length).toBeGreaterThan(0)
  })

  it('renders the nsfw switch alongside descriptive text', async () => {
    const wrapper = await mountSuspended(SensitiveContent)
    // The component has both a card title and an item title rendered by i18n
    // We assert there is at least one non-switch text element
    const spans = wrapper.findAll('span')
    const nonEmpty = spans.filter(s => s.text().trim().length > 0)
    expect(nonEmpty.length).toBeGreaterThan(0)
  })
})
