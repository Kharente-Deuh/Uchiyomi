// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { requiresHealthCheck, resolveHealthGuard } from './health-guard'
import { routeStub } from './route.stub'

describe('requiresHealthCheck', () => {
  it('exempte /status pour ne pas boucler sur la redirection', () => {
    expect(requiresHealthCheck(routeStub({ name: 'status' }))).toBe(false)
  })

  it('exige le check sur toute autre route', () => {
    expect(requiresHealthCheck(routeStub({ name: 'index' }))).toBe(true)
    expect(requiresHealthCheck(routeStub({ name: 'library' }))).toBe(true)
  })
})

describe('resolveHealthGuard', () => {
  const target = routeStub({ name: 'library', fullPath: '/library' })

  it('laisse passer un serveur prêt', () => {
    expect(resolveHealthGuard(target, { status: 'ok', components: {} })).toBeUndefined()
  })

  it('redirige un serveur en cours de démarrage', () => {
    expect(resolveHealthGuard(target, { status: 'starting', components: {} })).toBe('/status?redirect=%2Flibrary')
  })

  it('redirige un serveur en échec', () => {
    expect(resolveHealthGuard(target, { status: 'failed', components: {} })).toBe('/status?redirect=%2Flibrary')
  })

  it('redirige un serveur injoignable', () => {
    expect(resolveHealthGuard(target, { status: 'unreachable' })).toBe('/status?redirect=%2Flibrary')
  })

  it('encode la query string du chemin cible', () => {
    const paged = routeStub({ name: 'library', fullPath: '/library?page=2' })

    expect(resolveHealthGuard(paged, { status: 'starting', components: {} }))
      .toBe('/status?redirect=%2Flibrary%3Fpage%3D2')
  })
})
