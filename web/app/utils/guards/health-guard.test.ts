// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { requiresHealthCheck, resolveHealthGuard } from './health-guard'
import { routeStub } from './route.stub'

describe('requiresHealthCheck', () => {
  it('exempts /status to avoid redirect loops', () => {
    expect(requiresHealthCheck(routeStub({ name: 'status' }))).toBe(false)
  })

  it('requires the check on every other route', () => {
    expect(requiresHealthCheck(routeStub({ name: 'index' }))).toBe(true)
    expect(requiresHealthCheck(routeStub({ name: 'library' }))).toBe(true)
  })
})

describe('resolveHealthGuard', () => {
  const target = routeStub({ name: 'library', fullPath: '/library' })

  it('allows a ready server through', () => {
    expect(resolveHealthGuard(target, { status: 'ok', components: {} })).toBeUndefined()
  })

  it('redirects a starting server', () => {
    expect(resolveHealthGuard(target, { status: 'starting', components: {} })).toBe('/status?redirect=%2Flibrary')
  })

  it('redirects a failed server', () => {
    expect(resolveHealthGuard(target, { status: 'failed', components: {} })).toBe('/status?redirect=%2Flibrary')
  })

  it('redirects an unreachable server', () => {
    expect(resolveHealthGuard(target, { status: 'unreachable' })).toBe('/status?redirect=%2Flibrary')
  })

  it('encodes the target path query string', () => {
    const paged = routeStub({ name: 'library', fullPath: '/library?page=2' })

    expect(resolveHealthGuard(paged, { status: 'starting', components: {} }))
      .toBe('/status?redirect=%2Flibrary%3Fpage%3D2')
  })
})
