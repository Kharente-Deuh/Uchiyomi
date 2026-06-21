// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { resolveAuthGuard } from './auth-guard'

describe('resolveAuthGuard', () => {
  it('redirects unauthenticated users to /login', () => {
    expect(resolveAuthGuard('/settings', false)).toEqual({ redirect: '/login' })
  })
  it('lets authenticated users through', () => {
    expect(resolveAuthGuard('/settings', true)).toEqual({})
  })
  it('lets the public /login route pass for guests', () => {
    expect(resolveAuthGuard('/login', false)).toEqual({})
  })
  it('redirects authenticated users away from /login', () => {
    expect(resolveAuthGuard('/login', true)).toEqual({ redirect: '/' })
  })
})
