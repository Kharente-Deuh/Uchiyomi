// SPDX-License-Identifier: AGPL-3.0-or-later
import type { LoginRequestDto, SetupRequestDto, UserDto } from '#shared/dto/identity'
import type { ApiResponse } from '~/utils/api'
import { ApiError } from '~/utils/api'

export interface AuthApi {
  isSetupStatusRequired: () => Promise<ApiResponse<boolean>>
  setup: (body: SetupRequestDto) => Promise<ApiResponse<UserDto>>
  login: (body: LoginRequestDto) => Promise<ApiResponse<void>>
  me: () => Promise<ApiResponse<UserDto>>
  logout: () => Promise<ApiResponse<void>>
}

async function isSetupStatusRequired(): Promise<ApiResponse<boolean>> {
  try {
    const res = await $fetch('/api/auth/setup')

    return { success: true, data: res.required }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function setup(body: SetupRequestDto): Promise<ApiResponse<UserDto>> {
  try {
    const res = await $fetch('/api/auth/setup', { method: 'POST', body })

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function login(body: LoginRequestDto): Promise<ApiResponse<void>> {
  try {
    await $fetch('/api/auth/login', { method: 'POST', body })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function logout(): Promise<ApiResponse<void>> {
  try {
    await $fetch('/api/auth/logout', { method: 'POST' })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function me(): Promise<ApiResponse<UserDto>> {
  try {
    const res = await $fetch('/api/auth/me')

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

export function createAuthApi(): AuthApi {
  return {
    isSetupStatusRequired,
    setup,
    login,
    logout,
    me,
  }
}
