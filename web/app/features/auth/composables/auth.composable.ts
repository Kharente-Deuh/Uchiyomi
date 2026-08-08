// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AuthCheck } from '~/features/auth/store/auth.store'
import type { User } from '~/features/users/composables/users.api'

export interface AuthComposable {
  user: Ref<User | undefined>

  isAuthenticated: Ref<boolean>
  isAdmin: Ref<boolean>

  loading: Ref<boolean>

  fetchMe: () => Promise<AuthCheck>
  invalidate: () => void
}

export function useAuth(): AuthComposable {
  const store = useAuthStore()

  const user = computed(() => store.user)
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => !!user.value?.isAdmin)
  const loading = computed(() => store.status === 'loading')

  return {
    user,
    isAuthenticated,
    isAdmin,

    loading,

    fetchMe: store.fetchMe,
    invalidate: store.invalidate,
  }
}
