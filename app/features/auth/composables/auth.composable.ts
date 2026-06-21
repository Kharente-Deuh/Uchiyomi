import type { ComputedRef, Ref } from 'vue'
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { LoginRequestDto, SetupRequestDto, UserDto } from '#shared/dto/identity'
import type { ApiResponse } from '~/utils/api'
import { createAuthApi } from '../api'

const authApi = createAuthApi()

export interface AuthComposable {
  user: ComputedRef<UserDto | undefined>
  isAuthenticated: ComputedRef<boolean>
  loading: Ref<boolean>
  setupRequired: () => Promise<ApiResponse<boolean>>
  fetchMe: () => Promise<void>
  login: (body: LoginRequestDto) => Promise<ApiResponse<void>>
  setup: (body: SetupRequestDto) => Promise<ApiResponse<UserDto>>
  logout: () => Promise<ApiResponse<void>>
}

export function useAuth(): AuthComposable {
  const authStore = useAuthStore()
  const session = useUserSession()
  const loading = ref(false)

  // Hydrate the store from /me. A 401 means "not authenticated" → clear the
  // store and the session cookie. Other failures are not surfaced here (silent
  // boot probe); callers that need failure detail use the methods below.
  async function fetchMe(): Promise<void> {
    const res = await authApi.me()
    if (res.success) {
      authStore.setUser(res.data)

      return
    }

    if (res.error.status === 401) {
      authStore.clear()
      await session.clear()
    }
  }

  async function login(body: LoginRequestDto): Promise<ApiResponse<void>> {
    loading.value = true

    // login returns { ok } only → the user is hydrated via /me.
    const res = await authApi.login(body)
    if (res.success) {
      await fetchMe()
    }

    loading.value = false

    return res
  }

  async function setup(body: SetupRequestDto): Promise<ApiResponse<UserDto>> {
    loading.value = true

    const res = await authApi.setup(body)
    if (res.success) {
      authStore.setUser(res.data)
    }

    loading.value = false

    return res
  }

  async function logout(): Promise<ApiResponse<void>> {
    loading.value = true

    const res = await authApi.logout()
    if (res.success) {
      authStore.clear()
      await session.clear()
    }

    loading.value = false

    return res
  }

  return {
    user: computed(() => authStore.user),
    isAuthenticated: computed(() => authStore.isAuthenticated),
    loading,
    setupRequired: authApi.isSetupStatusRequired,
    fetchMe,
    login,
    setup,
    logout,
  }
}
