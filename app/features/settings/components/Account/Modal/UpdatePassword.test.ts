// SPDX-License-Identifier: AGPL-3.0-or-later

import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import UpdatePassword from './UpdatePassword.vue'

const { changePasswordMock } = vi.hoisted(() => ({ changePasswordMock: vi.fn() }))

function useAuthMock(): ReturnType<typeof useAuth> {
  return {
    loading: ref(false),
    minPasswordLength: ref(10),
    changePassword: changePasswordMock,
    user: ref(undefined),
    isAuthenticated: ref(false),
    needsAdmin: ref(false),
    getSetupStatus: vi.fn(),
    fetchMe: vi.fn(),
    login: vi.fn(),
    setup: vi.fn(),
    updateDisplayName: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>
}

mockNuxtImport('useAuth', () => useAuthMock)

async function fillAndSubmit(
  wrapper: Awaited<ReturnType<typeof mountSuspended<typeof UpdatePassword>>>,
  current: string,
  next: string,
  confirm: string,
  logoutOthers: boolean,
): Promise<void> {
  const state = (wrapper.vm as any).$.setupState
  state.field('currentPassword').handleChange(current)
  state.field('newPassword').handleChange(next)
  state.field('confirmNewPassword').handleChange(confirm)
  state.logoutOtherDevices = logoutOthers
  await flushPromises()
  await state.handleSubmit()
  await flushPromises()
}

describe('updatePassword modal', () => {
  afterEach(() => changePasswordMock.mockReset())

  it('submits the change-password request (with logoutOtherDevices) and closes on success', async () => {
    changePasswordMock.mockResolvedValue({ success: true, data: undefined })
    const wrapper = await mountSuspended(UpdatePassword, { props: { modelValue: true }, global: { stubs: { OrganismModal: { template: '<div><slot /></div>' } } } })
    await fillAndSubmit(wrapper, 'oldpass1234', 'newpass1234', 'newpass1234', true)
    expect(changePasswordMock).toHaveBeenCalledWith({
      currentPassword: 'oldpass1234',
      newPassword: 'newpass1234',
      logoutOtherDevices: true,
    })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([false])
  })

  it('does not submit when the new passwords do not match', async () => {
    const wrapper = await mountSuspended(UpdatePassword, { props: { modelValue: true }, global: { stubs: { OrganismModal: { template: '<div><slot /></div>' } } } })
    await fillAndSubmit(wrapper, 'oldpass1234', 'newpass1234', 'different999', false)
    expect(changePasswordMock).not.toHaveBeenCalled()
  })

  it('keeps the modal open on a failed change', async () => {
    changePasswordMock.mockResolvedValue({ success: false, error: { status: 400, message: 'x' } })
    const wrapper = await mountSuspended(UpdatePassword, { props: { modelValue: true }, global: { stubs: { OrganismModal: { template: '<div><slot /></div>' } } } })
    await fillAndSubmit(wrapper, 'oldpass1234', 'newpass1234', 'newpass1234', false)
    expect(changePasswordMock).toHaveBeenCalled()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
