// SPDX-License-Identifier: AGPL-3.0-or-later

import { DEFAULT_PAGE } from '~/constants'
import { isStatusPath } from './routes'

function isKnownRoute(path: string): boolean {
  const router = useRouter()

  return router.resolve(path).matched.length > 0
}

export function safeRedirect(raw: unknown): string {
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
