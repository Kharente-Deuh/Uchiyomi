// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ApiResponse } from '~/utils/api'
import { ApiError, apiFetch } from '~/utils/api'

export interface UsersApi {
  getCurrentUser: () => Promise<ApiResponse<User>>
}

export interface User {
  id: string
  username: string
  isAdmin: boolean
}

export function createUsersApi(): UsersApi {
  const api = apiFetch.create({ baseURL: '/api/users' })

  async function getCurrentUser(): Promise<ApiResponse<User>> {
    try {
      const user = await api<User>('/me')

      return { success: true, data: user }
    } catch (error) {
      return {
        success: false,
        error: ApiError.fromFetchError(error),
      }
    }
  }

  return { getCurrentUser }
}
