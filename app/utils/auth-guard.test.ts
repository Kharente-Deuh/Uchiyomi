// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { resolveAuthGuard } from './auth-guard'

const guest = { isAuthenticated: false, needsAdmin: false }
const auth = { isAuthenticated: true, needsAdmin: false }
const firstRun = { isAuthenticated: false, needsAdmin: true }

describe('resolveAuthGuard', () => {
  it('forces every route to /setup during first run', () => {
    expect(resolveAuthGuard('/', firstRun)).toEqual({ redirect: '/setup' })
    expect(resolveAuthGuard('/login', firstRun)).toEqual({ redirect: '/setup' })
  })
  it('allows /setup itself during first run', () => {
    expect(resolveAuthGuard('/setup', firstRun)).toEqual({})
  })
  it('redirects /setup away once setup is closed', () => {
    expect(resolveAuthGuard('/setup', guest)).toEqual({ redirect: '/login' })
    expect(resolveAuthGuard('/setup', auth)).toEqual({ redirect: '/' })
  })
  it('redirects unauthenticated users to /login with a redirect param', () => {
    expect(resolveAuthGuard('/settings', guest)).toEqual({ redirect: '/login?redirect=%2Fsettings' })
  })
  it('lets the bare /login route pass for guests', () => {
    expect(resolveAuthGuard('/login', guest)).toEqual({})
  })
  it('redirects authenticated users away from /login', () => {
    expect(resolveAuthGuard('/login', auth)).toEqual({ redirect: '/' })
  })
  it('lets authenticated users through guarded routes', () => {
    expect(resolveAuthGuard('/settings', auth)).toEqual({})
  })
})
