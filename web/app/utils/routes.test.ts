// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { routeStub } from './guards/route.stub'
import { isStatusPath, STATUS_PATH, statusRedirect } from './routes'

describe('isStatusPath', () => {
  it('matches the status path itself', () => {
    expect(isStatusPath(STATUS_PATH)).toBe(true)
  })

  it('is case-insensitive', () => {
    expect(isStatusPath('/STATUS')).toBe(true)
  })

  it('ignores trailing slashes', () => {
    expect(isStatusPath('/status//')).toBe(true)
  })

  it('ignores the query string', () => {
    expect(isStatusPath('/status?redirect=%2Fsettings')).toBe(true)
  })

  it('ignores the hash', () => {
    expect(isStatusPath('/status#components')).toBe(true)
  })

  it('rejects another path', () => {
    expect(isStatusPath('/settings')).toBe(false)
  })

  it('rejects a path merely prefixed by the status path', () => {
    expect(isStatusPath('/status-page')).toBe(false)
  })

  it('rejects a nested path under the status path', () => {
    expect(isStatusPath('/status/details')).toBe(false)
  })
})

describe('statusRedirect', () => {
  it('carries the current fullPath as an encoded redirect', () => {
    expect(statusRedirect(routeStub({ fullPath: '/settings/oidc?tab=1' })))
      .toBe('/status?redirect=%2Fsettings%2Foidc%3Ftab%3D1')
  })

  it('round-trips through decodeURIComponent', () => {
    const target = '/settings/oidc/1c9e#claims'
    const redirect = statusRedirect(routeStub({ fullPath: target })) as string
    expect(decodeURIComponent(redirect.split('redirect=', 2)[1] as string)).toBe(target)
  })
})
