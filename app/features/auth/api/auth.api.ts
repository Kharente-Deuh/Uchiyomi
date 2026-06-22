// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ChangePasswordRequestDto, LoginRequestDto, SetupRequestDto, SetupStatusDto, UpdateMeRequestDto, UserDto } from '#shared/dto/identity'
import type { ApiResponse } from '~/utils/api'
import { ApiError, apiFetch } from '~/utils/api'

export interface AuthApi {
  getSetupStatus: () => Promise<ApiResponse<SetupStatusDto>>
  setup: (body: SetupRequestDto) => Promise<ApiResponse<UserDto>>
  login: (body: LoginRequestDto) => Promise<ApiResponse<void>>
  me: () => Promise<ApiResponse<UserDto>>
  updateMe: (body: UpdateMeRequestDto) => Promise<ApiResponse<UserDto>>
  changePassword: (body: ChangePasswordRequestDto) => Promise<ApiResponse<void>>
  logout: () => Promise<ApiResponse<void>>
}

async function getSetupStatus(): Promise<ApiResponse<SetupStatusDto>> {
  try {
    const res = await apiFetch('/api/auth/setup')

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function setup(body: SetupRequestDto): Promise<ApiResponse<UserDto>> {
  try {
    const res = await apiFetch('/api/auth/setup', { method: 'POST', body })

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function login(body: LoginRequestDto): Promise<ApiResponse<void>> {
  try {
    await apiFetch('/api/auth/login', { method: 'POST', body })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function logout(): Promise<ApiResponse<void>> {
  try {
    await apiFetch('/api/auth/logout', { method: 'POST' })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function me(): Promise<ApiResponse<UserDto>> {
  try {
    const res = await apiFetch('/api/auth/me')

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function updateMe(body: UpdateMeRequestDto): Promise<ApiResponse<UserDto>> {
  try {
    const res = await apiFetch('/api/auth/me', { method: 'PATCH', body })

    return { success: true, data: res.user }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function changePassword(body: ChangePasswordRequestDto): Promise<ApiResponse<void>> {
  try {
    await apiFetch('/api/auth/me/password', { method: 'POST', body })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

export function createAuthApi(): AuthApi {
  return {
    getSetupStatus,
    setup,
    login,
    logout,
    me,
    updateMe,
    changePassword,
  }
}
