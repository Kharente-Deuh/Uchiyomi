// SPDX-License-Identifier: AGPL-3.0-or-later

import type { User } from '~/features/users/composables/users.api'
import { createUsersApi } from '~/features/users/composables/users.api'

export type AuthStatus = 'idle' | 'loading' | 'ready'

export type AuthCheck = 'ok' | 'unauthenticated' | 'unavailable'

export interface AuthStore {
  user: Ref<User | undefined>
  status: Ref<AuthStatus>

  fetchMe: () => Promise<AuthCheck>
  invalidate: () => void
}

export const useAuthStore = defineStore('auth', (): AuthStore => {
  const usersApi = createUsersApi()

  const user = ref<User>()
  const status = ref<AuthStatus>('idle')

  let pending: Promise<AuthCheck> | undefined

  function invalidate(): void {
    user.value = undefined
    status.value = 'idle'
  }

  async function load(): Promise<AuthCheck> {
    status.value = 'loading'

    const res = await usersApi.getCurrentUser()

    if (res.success) {
      user.value = res.data
      status.value = 'ready'

      return 'ok'
    }

    if (res.error.status === 401) {
      user.value = undefined
      status.value = 'ready'

      return 'unauthenticated'
    }

    invalidate()

    return 'unavailable'
  }

  async function fetchMe(): Promise<AuthCheck> {
    if (status.value === 'ready') {
      return user.value ? 'ok' : 'unauthenticated'
    }

    if (pending) {
      return pending
    }

    pending = load()

    try {
      return await pending
    } finally {
      pending = undefined
    }
  }

  return {
    user,
    status,

    fetchMe,
    invalidate,
  }
})
