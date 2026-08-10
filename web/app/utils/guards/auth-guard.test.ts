// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP, NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'
import { requiresAuthCheck, resolveAuthGuard } from './auth-guard'
import { routeStub } from './route.stub'

const ANONYMOUS = { isAuthenticated: false, isAdmin: false }
const USER = { isAuthenticated: true, isAdmin: false }
const ADMIN = { isAuthenticated: true, isAdmin: true }

describe('requiresAuthCheck', () => {
  it('skips a route without a group to spare the /me round-trip', () => {
    expect(requiresAuthCheck(routeStub({ name: 'status' }))).toBe(false)
    expect(requiresAuthCheck(routeStub({ name: 'status', meta: { authGroups: [] } }))).toBe(false)
  })

  it('requires the check as soon as a group is declared', () => {
    expect(requiresAuthCheck(routeStub({ meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] } }))).toBe(true)
  })
})

describe('resolveAuthGuard', () => {
  const library = routeStub({
    name: 'library',
    fullPath: '/library?page=2',
    meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] },
  })

  it('redirects anonymous user to /login, preserving the target for post-login', () => {
    expect(resolveAuthGuard({ to: library, ...ANONYMOUS })).toBe('/login?redirect=%2Flibrary%3Fpage%3D2')
  })

  it('redirects to /login without query from home, which is already the default target', () => {
    const index = routeStub({ name: 'index', fullPath: '/', meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] } })

    expect(resolveAuthGuard({ to: index, ...ANONYMOUS })).toBe('/login')
  })

  it('allows an authenticated user through', () => {
    expect(resolveAuthGuard({ to: library, ...USER })).toBeUndefined()
  })

  it('redirects an authenticated user on /login back to home', () => {
    const login = routeStub({ name: 'login', fullPath: '/login', meta: { authGroups: [NOT_AUTHENTICATED_ROUTE_GROUP] } })

    expect(resolveAuthGuard({ to: login, ...USER })).toBe('/')
    expect(resolveAuthGuard({ to: login, ...ANONYMOUS })).toBeUndefined()
  })
})

describe('resolveAuthGuard, admin group', () => {
  const admin = routeStub({ name: 'settings', fullPath: '/settings', meta: { authGroups: [ADMIN_ROUTE_GROUP] } })

  it('sends anonymous user to /login rather than bouncing through home', () => {
    expect(resolveAuthGuard({ to: admin, ...ANONYMOUS })).toBe('/login?redirect=%2Fsettings')
  })

  it('redirects authenticated user without rights to home', () => {
    expect(resolveAuthGuard({ to: admin, ...USER })).toBe('/')
  })

  it('allows an admin through', () => {
    expect(resolveAuthGuard({ to: admin, ...ADMIN })).toBeUndefined()
  })
})
