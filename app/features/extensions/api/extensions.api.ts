// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto, ExtensionHealthDto, ExtensionListQueryDto, ExtensionListResponseDto, PreferenceDto, SourceDto, UpdateSourcePreferenceRequestDto } from '#shared/dto/extensions'
import type { ApiResponse } from '~/utils/api'
import { ApiError, apiFetch } from '~/utils/api'

export interface ExtensionsApi {
  listExtensions: (params: ExtensionListQueryDto) => Promise<ApiResponse<ExtensionListResponseDto>>
  getExtension: (pkgName: string) => Promise<ApiResponse<{ extension: ExtensionDto, health: ExtensionHealthDto | null }>>
  extensionAction: (pkgName: string, action: 'install' | 'uninstall') => Promise<ApiResponse<void>>
  listSources: (pkgName: string) => Promise<ApiResponse<SourceDto[]>>
  setSourceEnabled: (sourceId: string, isEnabled: boolean) => Promise<ApiResponse<SourceDto>>
  getPreferences: (sourceId: string) => Promise<ApiResponse<PreferenceDto[]>>
  updatePreference: (sourceId: string, body: UpdateSourcePreferenceRequestDto) => Promise<ApiResponse<PreferenceDto[]>>
}

async function listExtensions(params: ExtensionListQueryDto): Promise<ApiResponse<ExtensionListResponseDto>> {
  try {
    const res = await apiFetch(`/api/extensions`, { query: params })

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function getExtension(pkgName: string): Promise<ApiResponse<{ extension: ExtensionDto, health: ExtensionHealthDto | null }>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}`)

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function extensionAction(pkgName: string, action: 'install' | 'uninstall'): Promise<ApiResponse<void>> {
  try {
    await apiFetch(`/api/extensions/${pkgName}`, { method: 'POST', body: { action } })

    return { success: true, data: undefined }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function listSources(pkgName: string): Promise<ApiResponse<SourceDto[]>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}/sources`)

    return { success: true, data: res.sources }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function setSourceEnabled(sourceId: string, isEnabled: boolean): Promise<ApiResponse<SourceDto>> {
  try {
    const res = await apiFetch(`/api/sources/${sourceId}`, { method: 'PATCH', body: { isEnabled } })

    return { success: true, data: res.source }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function getPreferences(sourceId: string): Promise<ApiResponse<PreferenceDto[]>> {
  try {
    const res = await apiFetch(`/api/sources/${sourceId}/preferences`)

    return { success: true, data: res.preferences }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function updatePreference(sourceId: string, body: UpdateSourcePreferenceRequestDto): Promise<ApiResponse<PreferenceDto[]>> {
  try {
    const res = await apiFetch(`/api/sources/${sourceId}/preferences`, { method: 'PUT', body })

    return { success: true, data: res.preferences }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

export function createExtensionsApi(): ExtensionsApi {
  return { listExtensions, getExtension, extensionAction, listSources, setSourceEnabled, getPreferences, updatePreference }
}
