// SPDX-License-Identifier: AGPL-3.0-or-later

import { initApi } from '~/utils/api'

export type ServerStatus = 'ok' | 'starting' | 'failed'

export type ServerStatusResponse = {
  status: ServerStatus
  components: Record<string, { status: ServerStatus, reason?: string }>
} | { status: 'unreachable' }

export interface HealthApi {
  getServerStatus: () => Promise<ServerStatusResponse>
}

const READYZ_TIMEOUT_MS = 5 * 1000

const serverStatuses = new Set(['ok', 'starting', 'failed', 'unreachable'])

function isServerStatusResponse(value: unknown): value is ServerStatusResponse {
  return typeof value === 'object'
    && value !== null
    && serverStatuses.has((value as { status?: unknown }).status as string)
}

async function getServerStatus(): Promise<ServerStatusResponse> {
  try {
    const res = await initApi().raw<ServerStatusResponse>('/readyz', {
      ignoreResponseError: true,
      timeout: READYZ_TIMEOUT_MS,
    })

    return isServerStatusResponse(res._data) ? res._data : { status: 'unreachable' }
  } catch {
    return { status: 'unreachable' }
  }
}

export function createHealthApi(): HealthApi {
  return {
    getServerStatus,
  }
}
