// SPDX-License-Identifier: AGPL-3.0-or-later

import { DEFAULT_PAGE } from '~/constants'
import { isStatusPath } from './routes'

export function safeRedirect(raw: unknown, isKnownRoute: (path: string) => boolean): string {
  const defaultRedirect = DEFAULT_PAGE as unknown as string

  if (typeof raw !== 'string') {
    return defaultRedirect
  }

  if (!raw.startsWith('/') || raw.startsWith('//') || raw.includes('\\')) {
    return defaultRedirect
  }

  if (isStatusPath(raw)) {
    return defaultRedirect
  }

  return isKnownRoute(raw) ? raw : defaultRedirect
}
