// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface OidcApi {
  getAll: () => Promise<ApiResponse<LightOidcProvider[]>>
  getById: (id: string) => Promise<ApiResponse<OidcProvider>>
  testByIssuerUrl: (url: string) => Promise<ApiResponse<TestResponse>>
  create: (request: CreateOidcProviderRequest) => Promise<ApiResponse<OidcProvider>>
  updateById: (id: string, request: CreateOidcProviderRequest) => Promise<ApiResponse<OidcProvider>>
  deleteById: (id: string) => Promise<ApiResponse<void>>
}

export interface CreateOidcProviderRequest {
  displayName: string
  issuerUrl: string
  clientId: string
  clientSecret: string
  usernameClaim: string
  scopes: string[]
  roleClaim?: string
  adminValues?: string[]
  allowedValues?: string[]
  autoProvision: boolean
}

export interface LightOidcProvider {
  id: string
  displayName: string
  issuerUrl: string
}

export type OidcProvider = Omit<CreateOidcProviderRequest, 'clientSecret'> & {
  id: string
  updatedAt: Date
  createdAt: Date
}

export interface TestResponse {
  issuer: string
  authorizationEndpoint: string
  tokenEndpoint: string
  userInfoEndpoint: string
  endSessionEndpoint: string
  redirectUri: string
  supportsRpInitiatedLogout: boolean
}

function createOidcProvider(data: OidcProvider): OidcProvider {
  return {
    id: data.id,
    displayName: data.displayName,
    issuerUrl: data.issuerUrl,
    clientId: data.clientId,
    usernameClaim: data.usernameClaim,
    scopes: data.scopes,
    roleClaim: data.roleClaim ?? undefined,
    adminValues: data.adminValues?.length ? data.adminValues : undefined,
    allowedValues: data.allowedValues?.length ? data.adminValues : undefined,
    autoProvision: data.autoProvision,
    updatedAt: new Date(data.createdAt),
    createdAt: new Date(data.updatedAt),
  }
}

export function createOidcApi(): OidcApi {
  const api = initApi('/oidc/providers')

  async function getAll(): Promise<ApiResponse<LightOidcProvider[]>> {
    try {
      const providers = await api<LightOidcProvider[]>('/')

      return { success: true, data: providers }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getById(id: string): Promise<ApiResponse<OidcProvider>> {
    try {
      const provider = await api<OidcProvider>(`/${id}`)

      return { success: true, data: provider }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function testByIssuerUrl(issuerUrl: string): Promise<ApiResponse<TestResponse>> {
    try {
      const test = await api<TestResponse>('/probe', { method: 'POST', body: { issuerUrl } })

      return { success: true, data: test }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function create({ roleClaim, adminValues, allowedValues, ...body }: CreateOidcProviderRequest): Promise<ApiResponse<OidcProvider>> {
    try {
      const provider = await api<OidcProvider>('/', {
        method: 'POST',
        body: {
          ...body,
          ...(roleClaim && {
            roleClaim,
            ...(!!adminValues?.length && { adminValues }),
            ...(!!allowedValues?.length && { allowedValues }),
          }),
        },
      })

      return { success: true, data: createOidcProvider(provider) }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function updateById(id: string, body: CreateOidcProviderRequest): Promise<ApiResponse<OidcProvider>> {
    try {
      const provider = await api<OidcProvider>('/', { method: 'PUT', body })

      return { success: true, data: createOidcProvider(provider) }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function deleteById(id: string): Promise<ApiResponse<void>> {
    try {
      await api<void>(`/${id}`, { method: 'DELETE' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    getAll,
    getById,
    testByIssuerUrl,
    create,
    updateById,
    deleteById,
  }
}
