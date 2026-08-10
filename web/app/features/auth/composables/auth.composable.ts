// SPDX-License-Identifier: AGPL-3.0-or-later

import type { LoginWithPwdRequest, LoginWithPwdStatus } from './auth.api'
import type { User } from '~/features/users/composables/users.api'
import { createAuthApi } from './auth.api'

export type AuthCheck = 'ok' | 'unauthenticated' | 'unreachable'

export interface AuthComposable {
  user: Ref<User | undefined>

  isAuthenticated: Ref<boolean>
  isAdmin: Ref<boolean>
  loading: Ref<boolean>

  fetchMe: () => Promise<AuthCheck>
  login: (request: LoginWithPwdRequest) => Promise<LoginWithPwdStatus>
  logout: () => Promise<string | undefined>
}

export function useAuth(): AuthComposable {
  const store = useAuthStore()
  const authApi = createAuthApi()
  const usersApi = createUsersApi()

  const user = computed(() => store.user)
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => !!user.value?.isAdmin)
  const loading = ref(false)

  async function login(request: LoginWithPwdRequest): Promise<LoginWithPwdStatus> {
    const res = await authApi.loginWithPwd(request)

    if (res.status === 'ok') {
      store.setUser(res.user)
    }

    return res.status
  }

  async function logout(): Promise<string | undefined> {
    loading.value = true

    const res = await authApi.logout()
    if (!res.success) {
      console.error('authApi.logout', res.error)
    }

    store.invalidate()

    loading.value = false

    return res.success ? res.data.endSessionUrl : undefined
  }

  async function fetchMe(): Promise<AuthCheck> {
    const res = await usersApi.getCurrentUser()
    if (res.success) {
      store.setUser(res.data)

      return 'ok'
    }

    store.invalidate()
    if (res.error.status === 401) {
      return 'unauthenticated'
    }

    return 'unreachable'
  }

  return {
    user,
    isAuthenticated,
    isAdmin,

    loading,

    login,
    fetchMe,
    logout,
  }
}
