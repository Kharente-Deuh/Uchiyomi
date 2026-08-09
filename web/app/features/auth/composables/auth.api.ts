// SPDX-License-Identifier: AGPL-3.0-or-later

import type { User } from '~/features/users/composables/users.api'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface AuthApi {
  loginWithPwd: (request: LoginWithPwdRequest) => Promise<LoginWithPwdResult>
  logout: () => Promise<ApiResponse<void>>
}

export type LoginWithPwdStatus = 'ok' | 'invalid-credentials' | 'unknown-error'

export type LoginWithPwdResult
  = | { status: 'ok', user: User }
    | { status: Exclude<LoginWithPwdStatus, 'ok'> }

export interface LoginWithPwdRequest {
  username: string
  password: string
}

export function createAuthApi(): AuthApi {
  const api = initApi('/auth')

  async function loginWithPwd(body: LoginWithPwdRequest): Promise<LoginWithPwdResult> {
    try {
      const user = await api<User>('/login', { method: 'POST', body })

      return { status: 'ok', user }
    } catch (error) {
      const apiError = ApiError.fromFetchError(error)
      if (apiError.status === 401) {
        return { status: 'invalid-credentials' }
      }

      console.error(loginWithPwd.name, apiError)

      return { status: 'unknown-error' }
    }
  }

  async function logout(): Promise<ApiResponse<void>> {
    try {
      await api('/logout', { method: 'POST' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    loginWithPwd,
    logout,
  }
}
