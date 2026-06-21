// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity/user.dto'

export interface AuthCapabilities {
  canManageExtensions: boolean
  canDownload: boolean
  allowNsfw: boolean
}

export interface AuthStore {
  user: Ref<UserDto | undefined>
  isAuthenticated: Ref<boolean>
  isAdmin: Ref<boolean>
  capabilities: Ref<AuthCapabilities>
  setUser: (value: UserDto) => void
  clear: () => void
}

export const useAuthStore = defineStore('auth', (): AuthStore => {
  const user = ref<UserDto>()
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')
  const capabilities = computed(() => ({
    canManageExtensions: user.value?.canManageExtensions ?? false,
    canDownload: user.value?.canDownload ?? false,
    allowNsfw: user.value?.allowNsfw ?? false,
  }))

  function setUser(value: UserDto): void {
    user.value = value
  }

  function clear(): void {
    user.value = undefined
  }

  return {
    user,
    isAuthenticated,
    isAdmin,
    capabilities,
    setUser,
    clear,
  }
})
