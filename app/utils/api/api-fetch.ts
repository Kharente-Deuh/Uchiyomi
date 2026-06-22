// SPDX-License-Identifier: AGPL-3.0-or-later
import type { FetchContext, FetchResponse } from 'ofetch'

type SessionLostHandler = () => void

let sessionLostHandler: SessionLostHandler | undefined

/** Register the handler invoked when an authenticated request unexpectedly 401s. */
export function onSessionLost(handler: SessionLostHandler): void {
  sessionLostHandler = handler
}

/** Test-only: clear the registered handler. */
export function __resetSessionLost(): void {
  sessionLostHandler = undefined
}

function requestUrl(request: FetchContext['request']): string {
  return typeof request === 'string' ? request : (request as Request).url
}

/**
 * Centralised 401 handling. A 401 from a non-`/api/auth/*` route is treated as a
 * lost session (expired or account disabled, M2.2 instant revocation) and delegated
 * to the registered handler, which decides what to do based on auth state. Auth
 * routes 401 legitimately (boot probe, bad credentials) and are ignored here.
 */
export function handleResponseError(
  ctx: FetchContext & { response: FetchResponse<unknown> },
): void {
  if (ctx.response?.status !== 401) {
    return
  }

  if (requestUrl(ctx.request).includes('/api/auth/')) {
    return
  }

  sessionLostHandler?.()
}

/** Shared fetch instance for the data-access layer; carries the 401 interceptor. */
export const apiFetch = $fetch.create({ onResponseError: handleResponseError })
