// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto, ExtensionListQueryDto, SourceDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import type { PageDto } from '#shared/dto/page.dto'
import type { ApiResponse } from '~/utils/api'
import { ApiError, apiFetch } from '~/utils/api'

export interface ExtensionsApi {
  listExtensions: (params: ExtensionListQueryDto) => Promise<ApiResponse<PageDto<ExtensionDto>>>
  getExtension: (pkgName: string) => Promise<ApiResponse<{ extension: ExtensionDto }>>
  extensionAction: (pkgName: string, action: 'install' | 'uninstall' | 'update') => Promise<ApiResponse<ExtensionDto>>
  listSources: (pkgName: string) => Promise<ApiResponse<SourceDto[]>>
  setSourceEnabled: (pkgName: string, sourceId: string, isEnabled: boolean) => Promise<ApiResponse<SourceDto>>
  getSettings: (pkgName: string) => Promise<ApiResponse<ExtensionSettingsDto>>
  updateSettings: (pkgName: string, body: Omit<ExtensionSettingsDto, 'pkgName'>) => Promise<ApiResponse<ExtensionSettingsDto>>
}

async function listExtensions({ search, isInstalled, hasUpdate, nsfw, page, pageSize }: ExtensionListQueryDto): Promise<ApiResponse<PageDto<ExtensionDto>>> {
  try {
    const res = await apiFetch(`/api/extensions`, {
      query: {
        page,
        pageSize,
        ...(!!search && { search }),
        ...(isInstalled !== undefined && { isInstalled }),
        ...(hasUpdate !== undefined && { hasUpdate }),
        ...(nsfw !== undefined && { nsfw }),
      },
    })

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function getExtension(pkgName: string): Promise<ApiResponse<{ extension: ExtensionDto }>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}`)

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function extensionAction(pkgName: string, action: 'install' | 'uninstall' | 'update'): Promise<ApiResponse<ExtensionDto>> {
  try {
    const extension = await apiFetch(`/api/extensions/${pkgName}`, { method: 'POST', body: { action } })

    return { success: true, data: extension }
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

async function setSourceEnabled(pkgName: string, sourceId: string, isEnabled: boolean): Promise<ApiResponse<SourceDto>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}/sources/${sourceId}/enable`, { method: 'POST', body: { isEnabled } })

    return { success: true, data: res.source }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function getSettings(pkgName: string): Promise<ApiResponse<ExtensionSettingsDto>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}/sources/settings`)

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

async function updateSettings(pkgName: string, body: Omit<ExtensionSettingsDto, 'pkgName'>): Promise<ApiResponse<ExtensionSettingsDto>> {
  try {
    const res = await apiFetch(`/api/extensions/${pkgName}/sources/settings`, { method: 'PUT', body })

    return { success: true, data: res }
  } catch (error) {
    return { success: false, error: ApiError.fromFetchError(error) }
  }
}

export function createExtensionsApi(): ExtensionsApi {
  return {
    listExtensions,
    getExtension,
    extensionAction,
    listSources,
    setSourceEnabled,
    getSettings,
    updateSettings,
  }
}
