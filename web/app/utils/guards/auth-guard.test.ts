// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP, NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'
import { requiresAuthCheck, resolveAuthGuard } from './auth-guard'
import { routeStub } from './route.stub'

const ANONYMOUS = { isAuthenticated: false, isAdmin: false }
const USER = { isAuthenticated: true, isAdmin: false }
const ADMIN = { isAuthenticated: true, isAdmin: true }

describe('requiresAuthCheck', () => {
  it('ignore une route sans groupe, pour lui épargner l\'aller-retour /me', () => {
    expect(requiresAuthCheck(routeStub({ name: 'status' }))).toBe(false)
    expect(requiresAuthCheck(routeStub({ name: 'status', meta: { authGroups: [] } }))).toBe(false)
  })

  it('exige le check dès qu\'un groupe est déclaré', () => {
    expect(requiresAuthCheck(routeStub({ meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] } }))).toBe(true)
  })
})

describe('resolveAuthGuard', () => {
  const library = routeStub({
    name: 'library',
    fullPath: '/library?page=2',
    meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] },
  })

  it('renvoie sur /login un anonyme, en gardant la cible pour l\'après-connexion', () => {
    expect(resolveAuthGuard({ to: library, ...ANONYMOUS })).toBe('/login?redirect=%2Flibrary%3Fpage%3D2')
  })

  it('renvoie sur /login sans query depuis l\'accueil, qui est déjà la cible par défaut', () => {
    const index = routeStub({ name: 'index', fullPath: '/', meta: { authGroups: [AUTHENTICATED_ROUTE_GROUP] } })

    expect(resolveAuthGuard({ to: index, ...ANONYMOUS })).toBe('/login')
  })

  it('laisse passer un utilisateur connecté', () => {
    expect(resolveAuthGuard({ to: library, ...USER })).toBeUndefined()
  })

  it('renvoie à l\'accueil un connecté qui revient sur /login', () => {
    const login = routeStub({ name: 'login', fullPath: '/login', meta: { authGroups: [NOT_AUTHENTICATED_ROUTE_GROUP] } })

    expect(resolveAuthGuard({ to: login, ...USER })).toBe('/')
    expect(resolveAuthGuard({ to: login, ...ANONYMOUS })).toBeUndefined()
  })
})

describe('resolveAuthGuard, groupe admin', () => {
  const admin = routeStub({ name: 'settings', fullPath: '/settings', meta: { authGroups: [ADMIN_ROUTE_GROUP] } })

  it('envoie un anonyme sur /login plutôt que de le faire rebondir par l\'accueil', () => {
    expect(resolveAuthGuard({ to: admin, ...ANONYMOUS })).toBe('/login?redirect=%2Fsettings')
  })

  it('renvoie à l\'accueil un connecté sans les droits', () => {
    expect(resolveAuthGuard({ to: admin, ...USER })).toBe('/')
  })

  it('laisse passer un admin', () => {
    expect(resolveAuthGuard({ to: admin, ...ADMIN })).toBeUndefined()
  })
})
