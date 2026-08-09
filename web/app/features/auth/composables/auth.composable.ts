// SPDX-License-Identifier: AGPL-3.0-or-later

import type { LoginWithPwdRequest, LoginWithPwdStatus } from './auth.api'
import type { AuthCheck } from '~/features/auth/store/auth.store'
import type { User } from '~/features/users/composables/users.api'
import { createAuthApi } from './auth.api'

export interface AuthComposable {
  user: Ref<User | undefined>

  isAuthenticated: Ref<boolean>
  isAdmin: Ref<boolean>
  loading: Ref<boolean>
  logoutLoading: Ref<boolean>

  fetchMe: () => Promise<AuthCheck>
  invalidate: () => void

  login: (request: LoginWithPwdRequest) => Promise<LoginWithPwdStatus>
  logout: () => Promise<void>
}

export function useAuth(): AuthComposable {
  const store = useAuthStore()
  const authApi = createAuthApi()

  const user = computed(() => store.user)
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => !!user.value?.isAdmin)
  const loading = computed(() => store.status === 'loading')
  const logoutLoading = ref(false)

  async function login(request: LoginWithPwdRequest): Promise<LoginWithPwdStatus> {
    const res = await authApi.loginWithPwd(request)

    if (res.status === 'ok') {
      store.setUser(res.user)
    }

    return res.status
  }

  async function logout(): Promise<void> {
    logoutLoading.value = true

    const res = await authApi.logout()
    if (!res.success) {
      console.error('authApi.logout', res.error)
    }

    store.invalidate()

    logoutLoading.value = false
  }

  return {
    user,
    isAuthenticated,
    isAdmin,

    loading,

    login,

    fetchMe: store.fetchMe,
    invalidate: store.invalidate,

    logout,
    logoutLoading,
  }
}
