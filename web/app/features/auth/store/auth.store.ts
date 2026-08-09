// SPDX-License-Identifier: AGPL-3.0-or-later

import type { User } from '~/features/users/composables/users.api'

export interface AuthStore {
  user: Ref<User | undefined>

  setUser: (user: User) => void
  invalidate: () => void
}

export const useAuthStore = defineStore('auth', (): AuthStore => {
  const user = ref<User>()

  function setUser(value: User): void {
    user.value = value
  }

  function invalidate(): void {
    user.value = undefined
  }

  return {
    user,

    setUser,
    invalidate,
  }
})
