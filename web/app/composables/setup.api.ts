// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ApiResponse } from '~/utils/api'
import { ApiError, apiFetch } from '~/utils/api'

export type DoSetupBody = {
  username: string
  password: string
}

export type SetupState = 'required' | 'done' | 'unknown'

interface SetupStatusResponse {
  required: boolean
}

export interface SetupApi {
  getSetupStatus: () => Promise<SetupState>
  doSetup: (body: DoSetupBody) => Promise<ApiResponse<void>>
}

function isSetupStatusResponse(value: unknown): value is SetupStatusResponse {
  return typeof value === 'object'
    && value !== null
    && typeof (value as { required?: unknown }).required === 'boolean'
}

export function createSetupApi(): SetupApi {
  const api = apiFetch.create({ baseURL: '/api/setup' })

  async function getSetupStatus(): Promise<SetupState> {
    try {
      const res = await api<SetupStatusResponse>('/status')

      if (!isSetupStatusResponse(res)) {
        return 'unknown'
      }

      return res.required ? 'required' : 'done'
    } catch {
      return 'unknown'
    }
  }

  async function doSetup(body: DoSetupBody): Promise<ApiResponse<void>> {
    try {
      await api('/', { method: 'POST', body })

      return {
        success: true,
        data: undefined,
      }
    } catch (error) {
      return {
        success: false,
        error: ApiError.fromFetchError(error),
      }
    }
  }

  return {
    getSetupStatus,
    doSetup,
  }
}
