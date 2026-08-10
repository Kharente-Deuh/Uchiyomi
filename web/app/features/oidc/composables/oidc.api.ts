// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface OidcApi {
  getAll: () => Promise<ApiResponse<LightOidcProvider[]>>
  getById: (id: string) => Promise<ApiResponse<OidcProviderDetails>>
  testByIssuerUrl: (url: string) => Promise<ApiResponse<TestResponse>>
  create: (request: CreateOidcProviderRequest) => Promise<ApiResponse<OidcProvider>>
  updateById: (id: string, request: UpdateOidcProviderRequest) => Promise<ApiResponse<OidcProvider>>
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

export type UpdateOidcProviderRequest = Omit<CreateOidcProviderRequest, 'clientSecret'>

export interface LightOidcProvider {
  id: string
  displayName: string
  createdAt: Date
  userCount: number
}

export type OidcProvider = Omit<CreateOidcProviderRequest, 'clientSecret'> & {
  id: string
  updatedAt: Date
  createdAt: Date
}

export type OidcProviderDetails = OidcProvider & {
  users: {
    id: string
    linkedAt: Date
    username: string
    isAdmin: boolean
  }[]
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

function createOidcProvider(p: OidcProvider): OidcProvider {
  const { adminValues, allowedValues, createdAt, updatedAt, roleClaim, ...data } = p

  return {
    ...data,
    roleClaim: p.roleClaim ?? undefined,
    adminValues: p.adminValues?.length ? p.adminValues : undefined,
    allowedValues: p.allowedValues?.length ? p.allowedValues : undefined,
    autoProvision: p.autoProvision,
    updatedAt: new Date(p.updatedAt),
    createdAt: new Date(p.createdAt),
  }
}

export function createOidcApi(): OidcApi {
  const api = initApi('/oidc/providers')

  async function getAll(): Promise<ApiResponse<LightOidcProvider[]>> {
    try {
      const providers = await api<LightOidcProvider[]>('/')

      return {
        success: true,
        data: providers.map(p => ({ ...p, createdAt: new Date(p.createdAt) })),
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getById(id: string): Promise<ApiResponse<OidcProviderDetails>> {
    try {
      const provider = await api<OidcProviderDetails>(`/${id}`)

      const { adminValues, allowedValues, createdAt, updatedAt, roleClaim, users, ...data } = provider

      return {
        success: true,
        data: {
          ...data,
          roleClaim: provider.roleClaim ?? undefined,
          adminValues: provider.adminValues?.length ? provider.adminValues : undefined,
          allowedValues: provider.allowedValues?.length ? provider.allowedValues : undefined,
          autoProvision: provider.autoProvision,
          updatedAt: new Date(provider.updatedAt),
          createdAt: new Date(provider.createdAt),
          users: users?.map(({ linkedAt, ...u }) => ({ ...u, linkedAt: new Date(linkedAt) })),
        },
      }
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

  async function updateById(id: string, { roleClaim, adminValues, allowedValues, ...body }: UpdateOidcProviderRequest): Promise<ApiResponse<OidcProvider>> {
    try {
      const provider = await api<OidcProvider>(`/${id}`, {
        method: 'PUT',
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
